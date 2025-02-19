package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	cs "github.com/cloudtrust/common-service"
	"github.com/cloudtrust/common-service/database"
	errorhandler "github.com/cloudtrust/common-service/errors"
	commonhttp "github.com/cloudtrust/common-service/http"
	"github.com/cloudtrust/common-service/idgenerator"
	"github.com/cloudtrust/common-service/log"
	"github.com/cloudtrust/common-service/metrics"
	"github.com/cloudtrust/common-service/middleware"
	"github.com/cloudtrust/common-service/security"
	"github.com/cloudtrust/common-service/tracing"
	"github.com/cloudtrust/common-service/tracking"
	"github.com/cloudtrust/keycloak-bridge/internal/keycloakb"
	"github.com/cloudtrust/keycloak-bridge/pkg/account"
	"github.com/cloudtrust/keycloak-bridge/pkg/event"
	"github.com/cloudtrust/keycloak-bridge/pkg/events"
	"github.com/cloudtrust/keycloak-bridge/pkg/export"
	"github.com/cloudtrust/keycloak-bridge/pkg/management"
	"github.com/cloudtrust/keycloak-bridge/pkg/statistics"
	keycloak "github.com/cloudtrust/keycloak-client"
	"github.com/go-kit/kit/endpoint"
	kit_log "github.com/go-kit/kit/log"
	kit_level "github.com/go-kit/kit/log/level"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// ComponentID is an unique ID generated at component startup.
	ComponentID = "unknown"
	// Environment is filled by the compiler.
	Environment = "unknown"
	// GitCommit is filled by the compiler.
	GitCommit = "unknown"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	ComponentID := strconv.FormatUint(rand.Uint64(), 10)

	// Logger.
	var logger = log.NewLeveledLogger(kit_log.NewJSONLogger(os.Stdout))
	{
		// Timestamp
		logger = log.With(logger, "_ts", kit_log.DefaultTimestampUTC)

		// Caller
		logger = log.With(logger, "caller", kit_log.Caller(6))

		// Add component name, component ID and version to the logger tags.
		logger = log.With(logger, "component_name", keycloakb.ComponentName, "component_id", ComponentID, "component_version", keycloakb.Version)
	}
	defer logger.Info("msg", "Shutdown")

	// Log component version infos.
	logger.Info("msg", "Starting")
	logger.Info("environment", Environment, "git_commit", GitCommit)

	// Configurations.
	var c = config(log.With(logger, "unit", "config"))
	var (
		// Component
		authorizationConfigFile = c.GetString("authorization-file")

		// Publishing
		httpAddrInternal   = c.GetString("internal-http-host-port")
		httpAddrManagement = c.GetString("management-http-host-port")
		httpAddrAccount    = c.GetString("account-http-host-port")

		// Keycloak
		keycloakConfig = keycloak.Config{
			AddrTokenProvider: c.GetString("keycloak-oidc-uri"),
			AddrAPI:           c.GetString("keycloak-api-uri"),
			Timeout:           c.GetDuration("keycloak-timeout"),
		}

		// Enabled units
		pprofRouteEnabled = c.GetBool("pprof-route-enabled")

		// Influx
		influxWriteInterval = c.GetDuration("influx-write-interval")

		// DB - for the moment used just for audit events
		auditRwDbParams = database.GetDbConfig(c, "db-audit-rw", !c.GetBool("events-db"))

		// DB - Read only user for audit events
		auditRoDbParams = database.GetDbConfig(c, "db-audit-ro", false)

		// DB for custom configuration
		configRwDbParams = database.GetDbConfig(c, "db-config-rw", !c.GetBool("config-db-rw"))
		configRoDbParams = database.GetDbConfig(c, "db-config-ro", !c.GetBool("config-db-ro"))

		// Rate limiting
		rateLimit = map[string]int{
			"event":      c.GetInt("rate-event"),
			"account":    c.GetInt("rate-account"),
			"management": c.GetInt("rate-management"),
			"statistics": c.GetInt("rate-statistics"),
			"events":     c.GetInt("rate-events"),
		}

		corsOptions = cors.Options{
			AllowedOrigins:   c.GetStringSlice("cors-allowed-origins"),
			AllowedMethods:   c.GetStringSlice("cors-allowed-methods"),
			AllowCredentials: c.GetBool("cors-allow-credential"),
			AllowedHeaders:   c.GetStringSlice("cors-allowed-headers"),
			ExposedHeaders:   c.GetStringSlice("cors-exposed-headers"),
			Debug:            c.GetBool("cors-debug"),
		}

		logLevel = c.GetString("log-level")
	)

	// Unique ID generator
	var idGenerator = idgenerator.New(keycloakb.ComponentName, ComponentID)

	// Critical errors channel.
	var errc = make(chan error)
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Configure log filtering
	{
		var level kit_level.Option
		var err error
		level, err = log.ConvertToLevel(logLevel)

		if err != nil {
			logger.Error("error", err)
			return
		}

		logger = log.AllowLevel(logger, level)
	}

	// Security - Audience required
	var audienceRequired string
	{
		audienceRequired = c.GetString("audience-required")

		if audienceRequired == "" {
			logger.Error("msg", "audience parameter(audience-required) cannot be empty")
			return
		}
	}

	// Security - Basic AuthN token to protect internal/event endpoint
	var eventExpectedAuthToken string
	{
		eventExpectedAuthToken = c.GetString("event-basic-auth-token")

		if eventExpectedAuthToken == "" {
			logger.Error("msg", "password for event endpoint (event-basic-auth-token) cannot be empty")
			return
		}
	}

	// Keycloak client.
	var keycloakClient *keycloak.Client
	{
		var err error
		keycloakClient, err = keycloak.New(keycloakConfig)

		if err != nil {
			logger.Error("msg", "could not create Keycloak client", "error", err)
			return
		}
	}

	// Keycloak adaptor for common-service library
	commonKcAdaptor := keycloakb.NewKeycloakAuthClient(keycloakClient, logger)

	// Authorization Manager
	var authorizationManager security.AuthorizationManager
	{
		var err error
		authorizationManager, err = security.NewAuthorizationManagerFromFile(commonKcAdaptor, logger, authorizationConfigFile)

		if err != nil {
			logger.Error("msg", "could not load authorizations", "error", err)
			return
		}
	}

	var sentryClient tracking.SentryTracking
	{
		var logger = log.With(logger, "unit", "sentry")
		var err error
		sentryClient, err = tracking.NewSentry(c, "sentry")
		if err != nil {
			logger.Error("msg", "could not create Sentry client", "error", err)
			return
		}
		defer sentryClient.Close()
	}

	var influxMetrics metrics.Metrics
	{
		var err error
		influxMetrics, err = metrics.NewMetrics(c, "influx", logger)
		if err != nil {
			logger.Error("msg", "could not create Influx client", "error", err)
			return
		}
		defer influxMetrics.Close()
	}

	// Jaeger client.
	var tracer tracing.OpentracingClient
	{
		var logger = log.With(logger, "unit", "jaeger")
		var err error

		tracer, err = tracing.CreateJaegerClient(c, "jaeger", keycloakb.ComponentName)
		if err != nil {
			logger.Error("msg", "could not create Jaeger tracer", "error", err)
			return
		}
		defer tracer.Close()
	}

	var eventsDBConn database.CloudtrustDB
	{
		var err error
		eventsDBConn, err = auditRwDbParams.OpenDatabase()
		if err != nil {
			logger.Error("msg", "could not create R/W DB connection for audit events", "error", err)
			return
		}
	}

	var eventsRODBConn database.CloudtrustDB
	{
		var err error
		eventsRODBConn, err = auditRoDbParams.OpenDatabase()
		if err != nil {
			logger.Error("msg", "could not create RO DB connection for audit events", "error", err)
			return
		}
	}

	var configurationRwDBConn database.CloudtrustDB
	{
		var err error
		configurationRwDBConn, err = configRwDbParams.OpenDatabase()
		if err != nil {
			logger.Error("msg", "could not create DB connection for configuration storage (RW)", "error", err)
			return
		}
	}

	var configurationRoDBConn database.CloudtrustDB
	{
		var err error
		configurationRoDBConn, err = configRoDbParams.OpenDatabase()
		if err != nil {
			logger.Error("msg", "could not create DB connection for configuration storage (RO)", "error", err)
			return
		}
	}

	// Event service.
	var eventEndpoints = event.Endpoints{}
	{
		var eventLogger = log.With(logger, "svc", "event")

		var consoleModule event.ConsoleModule
		{
			consoleModule = event.NewConsoleModule(log.With(eventLogger, "module", "console"))
			consoleModule = event.MakeConsoleModuleInstrumentingMW(influxMetrics.NewHistogram("console_module"))(consoleModule)
			consoleModule = event.MakeConsoleModuleLoggingMW(log.With(eventLogger, "mw", "module", "unit", "console"))(consoleModule)
			consoleModule = event.MakeConsoleModuleTracingMW(tracer)(consoleModule)
		}

		var statisticModule event.StatisticModule
		{
			statisticModule = event.NewStatisticModule(influxMetrics)
			statisticModule = event.MakeStatisticModuleInstrumentingMW(influxMetrics.NewHistogram("statistic_module"))(statisticModule)
			statisticModule = event.MakeStatisticModuleLoggingMW(log.With(eventLogger, "mw", "module", "unit", "statistic"))(statisticModule)
			statisticModule = event.MakeStatisticModuleTracingMW(tracer)(statisticModule)
		}

		// new module for sending the events to the DB
		var eventsDBModule database.EventsDBModule
		{
			eventsDBModule = database.NewEventsDBModule(eventsDBConn)
			eventsDBModule = event.MakeEventsDBModuleInstrumentingMW(influxMetrics.NewHistogram("eventsDB_module"))(eventsDBModule)
			eventsDBModule = event.MakeEventsDBModuleLoggingMW(log.With(eventLogger, "mw", "module", "unit", "eventsDB"))(eventsDBModule)
			eventsDBModule = event.MakeEventsDBModuleTracingMW(tracer)(eventsDBModule)
		}

		var eventAdminComponent event.AdminComponent
		{
			var fns = []event.FuncEvent{consoleModule.Print, statisticModule.Stats, eventsDBModule.Store}
			eventAdminComponent = event.NewAdminComponent(fns, fns, fns, fns)
			eventAdminComponent = event.MakeAdminComponentInstrumentingMW(influxMetrics.NewHistogram("admin_component"))(eventAdminComponent)
			eventAdminComponent = event.MakeAdminComponentLoggingMW(log.With(eventLogger, "mw", "component", "unit", "admin_event"))(eventAdminComponent)
			eventAdminComponent = event.MakeAdminComponentTracingMW(tracer)(eventAdminComponent)
		}

		var eventComponent event.Component
		{
			var fns = []event.FuncEvent{consoleModule.Print, statisticModule.Stats, eventsDBModule.Store}
			eventComponent = event.NewComponent(fns, fns)
			eventComponent = event.MakeComponentInstrumentingMW(influxMetrics.NewHistogram("component"))(eventComponent)
			eventComponent = event.MakeComponentLoggingMW(log.With(eventLogger, "mw", "component", "unit", "event"))(eventComponent)
			eventComponent = event.MakeComponentTracingMW(tracer)(eventComponent)
		}

		// add ct_type

		var muxComponent event.MuxComponent
		{
			muxComponent = event.NewMuxComponent(eventComponent, eventAdminComponent)
			muxComponent = event.MakeMuxComponentInstrumentingMW(influxMetrics.NewHistogram("mux_component"))(muxComponent)
			muxComponent = event.MakeMuxComponentLoggingMW(log.With(eventLogger, "mw", "component", "unit", "mux"))(muxComponent)
			muxComponent = event.MakeMuxComponentTracingMW(tracer)(muxComponent)
			muxComponent = event.MakeMuxComponentTrackingMW(sentryClient, log.With(eventLogger, "mw", "component"))(muxComponent)
		}

		var eventEndpoint cs.Endpoint
		{
			eventEndpoint = event.MakeEventEndpoint(muxComponent)
			eventEndpoint = middleware.MakeEndpointInstrumentingMW(influxMetrics, "event_endpoint")(eventEndpoint)
			eventEndpoint = middleware.MakeEndpointLoggingMW(log.With(eventLogger, "mw", "endpoint"))(eventEndpoint)
			eventEndpoint = tracer.MakeEndpointTracingMW("event_endpoint")(eventEndpoint)
		}

		eventEndpoints = event.Endpoints{
			Endpoint: keycloakb.LimitRate(eventEndpoint, rateLimit["event"]),
		}
	}

	baseEventsDBModule := database.NewEventsDBModule(eventsDBConn)

	// new module for reading events from the DB
	eventsRODBModule := keycloakb.NewEventsDBModule(eventsRODBConn)

	// Statistics service.
	var statisticsEndpoints statistics.Endpoints
	{
		var statisticsLogger = log.With(logger, "svc", "statistics")

		statisticsComponent := statistics.NewComponent(eventsRODBModule, keycloakClient, statisticsLogger)
		statisticsComponent = statistics.MakeAuthorizationManagementComponentMW(log.With(statisticsLogger, "mw", "endpoint"), authorizationManager)(statisticsComponent)

		statisticsEndpoints = statistics.Endpoints{
			GetStatistics:      prepareEndpoint(statistics.MakeGetStatisticsEndpoint(statisticsComponent), "get_statistics", influxMetrics, statisticsLogger, tracer, rateLimit["statistics"]),
			GetMigrationReport: prepareEndpoint(statistics.MakeGetMigrationReportEndpoint(statisticsComponent), "get_migration_report", influxMetrics, statisticsLogger, tracer, rateLimit["statistics"]),
		}
	}

	// Events service.
	var eventsEndpoints events.Endpoints
	{
		var eventsLogger = log.With(logger, "svc", "events")

		// module to store API calls of the back office to the DB
		eventsDBModule := configureEventsDbModule(baseEventsDBModule, influxMetrics, eventsLogger, tracer)

		eventsComponent := events.NewComponent(eventsRODBModule, eventsDBModule, eventsLogger)
		eventsComponent = events.MakeAuthorizationManagementComponentMW(log.With(eventsLogger, "mw", "endpoint"), authorizationManager)(eventsComponent)

		eventsEndpoints = events.Endpoints{
			GetEvents:        prepareEndpoint(events.MakeGetEventsEndpoint(eventsComponent), "get_events", influxMetrics, eventsLogger, tracer, rateLimit["events"]),
			GetEventsSummary: prepareEndpoint(events.MakeGetEventsSummaryEndpoint(eventsComponent), "get_events_summary", influxMetrics, eventsLogger, tracer, rateLimit["events"]),
			GetUserEvents:    prepareEndpoint(events.MakeGetUserEventsEndpoint(eventsComponent), "get_user_events", influxMetrics, eventsLogger, tracer, rateLimit["events"]),
		}
	}

	// Management service.
	var managementEndpoints = management.Endpoints{}
	{
		var managementLogger = log.With(logger, "svc", "management")

		// module to store API calls of the back office to the DB
		eventsDBModule := configureEventsDbModule(baseEventsDBModule, influxMetrics, managementLogger, tracer)

		// module for storing and retrieving the custom configuration
		var configDBModule management.ConfigurationDBModule
		{
			configDBModule = keycloakb.NewConfigurationDBModule(configurationRwDBConn)
			configDBModule = keycloakb.MakeConfigurationDBModuleInstrumentingMW(influxMetrics.NewHistogram("configDB_module"))(configDBModule)
			configDBModule = management.MakeConfigurationDBModuleLoggingMW(log.With(managementLogger, "mw", "module", "unit", "configDB"))(configDBModule)
			configDBModule = management.MakeConfigurationDBModuleTracingMW(tracer)(configDBModule)
		}

		var keycloakComponent management.Component
		{
			keycloakComponent = management.NewComponent(keycloakClient, eventsDBModule, configDBModule, managementLogger)
			keycloakComponent = management.MakeAuthorizationManagementComponentMW(log.With(managementLogger, "mw", "endpoint"), authorizationManager)(keycloakComponent)
		}

		managementEndpoints = management.Endpoints{
			GetRealms:                      prepareEndpoint(management.MakeGetRealmsEndpoint(keycloakComponent), "realms_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetRealm:                       prepareEndpoint(management.MakeGetRealmEndpoint(keycloakComponent), "realm_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetClients:                     prepareEndpoint(management.MakeGetClientsEndpoint(keycloakComponent), "get_clients_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetClient:                      prepareEndpoint(management.MakeGetClientEndpoint(keycloakComponent), "get_client_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			CreateUser:                     prepareEndpoint(management.MakeCreateUserEndpoint(keycloakComponent), "create_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetUser:                        prepareEndpoint(management.MakeGetUserEndpoint(keycloakComponent), "get_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			UpdateUser:                     prepareEndpoint(management.MakeUpdateUserEndpoint(keycloakComponent), "update_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			DeleteUser:                     prepareEndpoint(management.MakeDeleteUserEndpoint(keycloakComponent), "delete_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetUsers:                       prepareEndpoint(management.MakeGetUsersEndpoint(keycloakComponent), "get_users_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetUserAccountStatus:           prepareEndpoint(management.MakeGetUserAccountStatusEndpoint(keycloakComponent), "get_user_accountstatus", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetGroupsOfUser:                prepareEndpoint(management.MakeGetGroupsOfUserEndpoint(keycloakComponent), "get_user_groups", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetRolesOfUser:                 prepareEndpoint(management.MakeGetRolesOfUserEndpoint(keycloakComponent), "get_user_roles", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetRoles:                       prepareEndpoint(management.MakeGetRolesEndpoint(keycloakComponent), "get_roles_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetRole:                        prepareEndpoint(management.MakeGetRoleEndpoint(keycloakComponent), "get_role_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetGroups:                      prepareEndpoint(management.MakeGetGroupsEndpoint(keycloakComponent), "get_groups_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetClientRoles:                 prepareEndpoint(management.MakeGetClientRolesEndpoint(keycloakComponent), "get_client_roles_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			CreateClientRole:               prepareEndpoint(management.MakeCreateClientRoleEndpoint(keycloakComponent), "create_client_role_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetClientRoleForUser:           prepareEndpoint(management.MakeGetClientRolesForUserEndpoint(keycloakComponent), "get_client_roles_for_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			AddClientRoleToUser:            prepareEndpoint(management.MakeAddClientRolesToUserEndpoint(keycloakComponent), "get_client_roles_for_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			ResetPassword:                  prepareEndpoint(management.MakeResetPasswordEndpoint(keycloakComponent), "reset_password_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			SendVerifyEmail:                prepareEndpoint(management.MakeSendVerifyEmailEndpoint(keycloakComponent), "send_verify_email_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			ExecuteActionsEmail:            prepareEndpoint(management.MakeExecuteActionsEmailEndpoint(keycloakComponent), "execute_actions_email_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			SendReminderEmail:              prepareEndpoint(management.MakeSendReminderEmailEndpoint(keycloakComponent), "send_reminder_email_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			SendNewEnrolmentCode:           prepareEndpoint(management.MakeSendNewEnrolmentCodeEndpoint(keycloakComponent), "send_new_enrolment_code_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetCredentialsForUser:          prepareEndpoint(management.MakeGetCredentialsForUserEndpoint(keycloakComponent), "get_credentials_for_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			DeleteCredentialsForUser:       prepareEndpoint(management.MakeDeleteCredentialsForUserEndpoint(keycloakComponent), "delete_credentials_for_user_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			GetRealmCustomConfiguration:    prepareEndpoint(management.MakeGetRealmCustomConfigurationEndpoint(keycloakComponent), "get_realm_custom_config_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
			UpdateRealmCustomConfiguration: prepareEndpoint(management.MakeUpdateRealmCustomConfigurationEndpoint(keycloakComponent), "update_realm_custom_config_endpoint", influxMetrics, managementLogger, tracer, rateLimit["management"]),
		}
	}

	// Account service.
	var accountEndpoints account.Endpoints
	{
		var accountLogger = log.With(logger, "svc", "account")

		// Configure events db module
		eventsDBModule := configureEventsDbModule(baseEventsDBModule, influxMetrics, accountLogger, tracer)

		// module for retrieving the custom configuration
		var configDBModule account.ConfigurationDBModule
		{
			configDBModule = keycloakb.NewConfigurationDBModule(configurationRoDBConn)
			configDBModule = keycloakb.MakeConfigurationDBModuleInstrumentingMW(influxMetrics.NewHistogram("configDB_module"))(configDBModule)
			configDBModule = management.MakeConfigurationDBModuleLoggingMW(log.With(logger, "mw", "module", "unit", "configDB"))(configDBModule)
			configDBModule = management.MakeConfigurationDBModuleTracingMW(tracer)(configDBModule)
		}

		// new module for account service
		accountComponent := account.NewComponent(keycloakClient.AccountClient(), eventsDBModule, configDBModule, accountLogger)
		accountComponent = account.MakeAuthorizationAccountComponentMW(log.With(accountLogger, "mw", "endpoint"), configDBModule)(accountComponent)

		accountEndpoints = account.Endpoints{
			GetAccount:                prepareEndpoint(account.MakeGetAccountEndpoint(accountComponent), "get_account", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			UpdateAccount:             prepareEndpoint(account.MakeUpdateAccountEndpoint(accountComponent), "update_account", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			DeleteAccount:             prepareEndpoint(account.MakeDeleteAccountEndpoint(accountComponent), "delete_account", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			UpdatePassword:            prepareEndpoint(account.MakeUpdatePasswordEndpoint(accountComponent), "update_password", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			GetCredentials:            prepareEndpoint(account.MakeGetCredentialsEndpoint(accountComponent), "get_credentials", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			GetCredentialRegistrators: prepareEndpoint(account.MakeGetCredentialRegistratorsEndpoint(accountComponent), "get_credential_registrators", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			DeleteCredential:          prepareEndpoint(account.MakeDeleteCredentialEndpoint(accountComponent), "delete_credential", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			UpdateLabelCredential:     prepareEndpoint(account.MakeUpdateLabelCredentialEndpoint(accountComponent), "update_label_credential", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			MoveCredential:            prepareEndpoint(account.MakeMoveCredentialEndpoint(accountComponent), "move_credential", influxMetrics, accountLogger, tracer, rateLimit["account"]),
			GetConfiguration:          prepareEndpoint(account.MakeGetConfigurationEndpoint(accountComponent), "get_configuration", influxMetrics, accountLogger, tracer, rateLimit["account"]),
		}
	}

	// Export configuration
	var exportModule = export.NewModule(keycloakClient, logger)
	var cfgStorageModue = export.NewConfigStorageModule(eventsDBConn)

	var exportComponent = export.NewComponent(keycloakb.ComponentName, keycloakb.Version, logger, exportModule, cfgStorageModue)
	var exportEndpoint = export.MakeExportEndpoint(exportComponent)
	var exportSaveAndExportEndpoint = export.MakeStoreAndExportEndpoint(exportComponent)

	errorhandler.SetEmitter(keycloakb.ComponentName)

	// HTTP Internal Call Server (Event reception from Keycloak & Export API).
	go func() {
		var logger = log.With(logger, "transport", "http")
		logger.Info("addr", httpAddrInternal)

		var route = mux.NewRouter()

		// Version.
		route.Handle("/", commonhttp.MakeVersionHandler(keycloakb.ComponentName, ComponentID, keycloakb.Version, Environment, GitCommit))

		// Event.
		var eventSubroute = route.PathPrefix("/event").Subrouter()

		var eventHandler http.Handler
		{
			eventHandler = event.MakeHTTPEventHandler(eventEndpoints.Endpoint, logger)
			eventHandler = middleware.MakeHTTPCorrelationIDMW(idGenerator, tracer, logger, keycloakb.ComponentName, ComponentID)(eventHandler)
			eventHandler = tracer.MakeHTTPTracingMW(keycloakb.ComponentName, "http_server_event")(eventHandler)
			eventHandler = middleware.MakeHTTPBasicAuthenticationMW(eventExpectedAuthToken, logger)(eventHandler)
		}
		eventSubroute.Handle("/receiver", eventHandler)

		// Export.
		route.Handle("/export", export.MakeHTTPExportHandler(exportEndpoint)).Methods("GET")
		route.Handle("/export", export.MakeHTTPExportHandler(exportSaveAndExportEndpoint)).Methods("POST")

		// Debug.
		if pprofRouteEnabled {
			var debugSubroute = route.PathPrefix("/debug").Subrouter()
			debugSubroute.HandleFunc("/pprof/", http.HandlerFunc(pprof.Index))
			debugSubroute.HandleFunc("/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
			debugSubroute.HandleFunc("/pprof/profile", http.HandlerFunc(pprof.Profile))
			debugSubroute.HandleFunc("/pprof/symbol", http.HandlerFunc(pprof.Symbol))
			debugSubroute.HandleFunc("/pprof/trace", http.HandlerFunc(pprof.Trace))
		}

		errc <- http.ListenAndServe(httpAddrInternal, route)
	}()

	// HTTP Management Server (Backoffice API).
	go func() {
		var logger = log.With(logger, "transport", "http")
		logger.Info("addr", httpAddrManagement)

		var route = mux.NewRouter()

		// Version.
		route.Handle("/", http.HandlerFunc(commonhttp.MakeVersionHandler(keycloakb.ComponentName, ComponentID, keycloakb.Version, Environment, GitCommit)))

		// Rights
		var rightsHandler = configureRightsHandler(keycloakb.ComponentName, ComponentID, idGenerator, authorizationManager, keycloakClient, audienceRequired, tracer, logger)
		route.Path("/rights").Methods("GET").Handler(rightsHandler)

		// Statistics
		var getStatisticsHandler = configureStatisiticsHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(statisticsEndpoints.GetStatistics)
		var getMigrationReportHandler = configureStatisiticsHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(statisticsEndpoints.GetMigrationReport)
		route.Path("/statistics/realms/{realm}").Methods("GET").Handler(getStatisticsHandler)
		route.Path("/statistics/realms/{realm}/migration").Methods("GET").Handler(getMigrationReportHandler)

		// Events
		var getEventsHandler = configureEventsHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(eventsEndpoints.GetEvents)
		var getEventsSummaryHandler = configureEventsHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(eventsEndpoints.GetEventsSummary)
		var getUserEventsHandler = configureEventsHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(eventsEndpoints.GetUserEvents)

		route.Path("/events").Methods("GET").Handler(getEventsHandler)
		route.Path("/events/summary").Methods("GET").Handler(getEventsSummaryHandler)
		route.Path("/events/realms/{realm}/users/{userID}/events").Methods("GET").Handler(getUserEventsHandler)

		// Management
		var managementSubroute = route.PathPrefix("/management").Subrouter()

		var getRealmsHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetRealms)
		var getRealmHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetRealm)

		var getClientsHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetClients)
		var getClientHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetClient)

		var createUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.CreateUser)
		var getUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetUser)
		var updateUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.UpdateUser)
		var deleteUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.DeleteUser)
		var getUsersHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetUsers)
		var getRolesForUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetRolesOfUser)
		var getGroupsForUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetGroupsOfUser)
		var getUserAccountStatusHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetUserAccountStatus)

		var getClientRoleForUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetClientRoleForUser)
		var addClientRoleToUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.AddClientRoleToUser)

		var getRolesHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetRoles)
		var getRoleHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetRole)
		var getClientRolesHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetClientRoles)
		var createClientRolesHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.CreateClientRole)

		var getGroupsHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetGroups)

		var resetPasswordHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.ResetPassword)
		var sendVerifyEmailHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.SendVerifyEmail)
		var executeActionsEmailHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.ExecuteActionsEmail)
		var sendNewEnrolmentCodeHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.SendNewEnrolmentCode)
		var sendReminderEmailHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.SendReminderEmail)

		var getCredentialsForUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetCredentialsForUser)
		var deleteCredentialsForUserHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.DeleteCredentialsForUser)

		var getRealmCustomConfigurationHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.GetRealmCustomConfiguration)
		var updateRealmCustomConfigurationHandler = configureManagementHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(managementEndpoints.UpdateRealmCustomConfiguration)

		//realms
		managementSubroute.Path("/realms").Methods("GET").Handler(getRealmsHandler)
		managementSubroute.Path("/realms/{realm}").Methods("GET").Handler(getRealmHandler)

		//clients
		managementSubroute.Path("/realms/{realm}/clients").Methods("GET").Handler(getClientsHandler)
		managementSubroute.Path("/realms/{realm}/clients/{clientID}").Methods("GET").Handler(getClientHandler)

		//users
		managementSubroute.Path("/realms/{realm}/users").Methods("GET").Handler(getUsersHandler)
		managementSubroute.Path("/realms/{realm}/users").Methods("POST").Handler(createUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}").Methods("GET").Handler(getUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}").Methods("PUT").Handler(updateUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}").Methods("DELETE").Handler(deleteUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/groups").Methods("GET").Handler(getGroupsForUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/roles").Methods("GET").Handler(getRolesForUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/status").Methods("GET").Handler(getUserAccountStatusHandler)

		//role mappings
		managementSubroute.Path("/realms/{realm}/users/{userID}/role-mappings/clients/{clientID}").Methods("GET").Handler(getClientRoleForUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/role-mappings/clients/{clientID}").Methods("POST").Handler(addClientRoleToUserHandler)

		managementSubroute.Path("/realms/{realm}/users/{userID}/reset-password").Methods("PUT").Handler(resetPasswordHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/send-verify-email").Methods("PUT").Handler(sendVerifyEmailHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/execute-actions-email").Methods("PUT").Handler(executeActionsEmailHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/send-new-enrolment-code").Methods("POST").Handler(sendNewEnrolmentCodeHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/send-reminder-email").Methods("POST").Handler(sendReminderEmailHandler)

		// Credentials
		managementSubroute.Path("/realms/{realm}/users/{userID}/credentials").Methods("GET").Handler(getCredentialsForUserHandler)
		managementSubroute.Path("/realms/{realm}/users/{userID}/credentials/{credentialID}").Methods("DELETE").Handler(deleteCredentialsForUserHandler)

		//roles
		managementSubroute.Path("/realms/{realm}/roles").Methods("GET").Handler(getRolesHandler)
		managementSubroute.Path("/realms/{realm}/roles-by-id/{roleID}").Methods("GET").Handler(getRoleHandler)
		managementSubroute.Path("/realms/{realm}/clients/{clientID}/roles").Methods("GET").Handler(getClientRolesHandler)
		managementSubroute.Path("/realms/{realm}/clients/{clientID}/roles").Methods("POST").Handler(createClientRolesHandler)

		//groups
		managementSubroute.Path("/realms/{realm}/groups").Methods("GET").Handler(getGroupsHandler)

		// custom configuration par realm
		managementSubroute.Path("/realms/{realm}/configuration").Methods("GET").Handler(getRealmCustomConfigurationHandler)
		managementSubroute.Path("/realms/{realm}/configuration").Methods("PUT").Handler(updateRealmCustomConfigurationHandler)

		c := cors.New(corsOptions)
		errc <- http.ListenAndServe(httpAddrManagement, c.Handler(route))

	}()

	// HTTP Self-service Server (Account API).
	go func() {
		var logger = log.With(logger, "transport", "http")
		logger.Info("addr", httpAddrAccount)

		var route = mux.NewRouter()

		// Version.
		route.Handle("/", http.HandlerFunc(commonhttp.MakeVersionHandler(keycloakb.ComponentName, ComponentID, keycloakb.Version, Environment, GitCommit)))

		// Account
		var updatePasswordHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.UpdatePassword)
		var getCredentialsHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.GetCredentials)
		var getCredentialRegistratorsHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.GetCredentialRegistrators)
		var deleteCredentialHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.DeleteCredential)
		var updateLabelCredentialHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.UpdateLabelCredential)
		var moveCredentialHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.MoveCredential)
		var getAccountHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.GetAccount)
		var updateAccountHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.UpdateAccount)
		var deleteAccountHandler = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.DeleteAccount)
		var getGetConfiguration = configureAccountHandler(keycloakb.ComponentName, ComponentID, idGenerator, keycloakClient, audienceRequired, tracer, logger)(accountEndpoints.GetConfiguration)

		route.Path("/account").Methods("GET").Handler(getAccountHandler)
		route.Path("/account").Methods("POST").Handler(updateAccountHandler)
		route.Path("/account").Methods("DELETE").Handler(deleteAccountHandler)

		route.Path("/account/configuration").Methods("GET").Handler(getGetConfiguration)

		route.Path("/account/credentials").Methods("GET").Handler(getCredentialsHandler)
		route.Path("/account/credentials/password").Methods("POST").Handler(updatePasswordHandler)
		route.Path("/account/credentials/registrators").Methods("GET").Handler(getCredentialRegistratorsHandler)
		route.Path("/account/credentials/{credentialID}").Methods("DELETE").Handler(deleteCredentialHandler)
		route.Path("/account/credentials/{credentialID}").Methods("PUT").Handler(updateLabelCredentialHandler)
		route.Path("/account/credentials/{credentialID}/after/{previousCredentialID}").Methods("POST").Handler(moveCredentialHandler)

		c := cors.New(corsOptions)
		errc <- http.ListenAndServe(httpAddrAccount, c.Handler(route))
	}()

	// Influx writing.
	go func() {
		var tic = time.NewTicker(influxWriteInterval)
		defer tic.Stop()
		influxMetrics.WriteLoop(tic.C)
	}()

	logger.Info("msg", "Started")
	logger.Error("error", <-errc)
}

func config(logger log.Logger) *viper.Viper {
	logger.Info("msg", "load configuration and command args")

	var v = viper.New()

	// Component default.
	v.SetDefault("config-file", "./configs/keycloak_bridge.yml")
	v.SetDefault("authorization-file", "./configs/authorization.json")

	// Log level
	v.SetDefault("log-level", "info")

	// Publishing
	v.SetDefault("internal-http-host-port", "0.0.0.0:8888")
	v.SetDefault("management-http-host-port", "0.0.0.0:8877")
	v.SetDefault("account-http-host-port", "0.0.0.0:8866")

	// Security - Audience check
	v.SetDefault("audience-required", "")
	v.SetDefault("event-basic-auth-token", "")

	// CORS configuration
	v.SetDefault("cors-allowed-origins", []string{})
	v.SetDefault("cors-allowed-methods", []string{})
	v.SetDefault("cors-allow-credentials", true)
	v.SetDefault("cors-allowed-headers", []string{})
	v.SetDefault("cors-exposed-headers", []string{})
	v.SetDefault("cors-debug", false)

	// Keycloak default.
	v.SetDefault("keycloak-api-uri", "http://127.0.0.1:8080")
	v.SetDefault("keycloak-oidc-uri", "http://127.0.0.1:8080 http://localhost:8080")
	v.SetDefault("keycloak-timeout", "5s")

	// Storage events in DB (read/write)
	v.SetDefault("events-db", false)
	database.ConfigureDbDefault(v, "db-audit-rw", "CT_BRIDGE_DB_AUDIT_RW_USERNAME", "CT_BRIDGE_DB_AUDIT_RW_PASSWORD")

	// Storage events in DB (read only)
	database.ConfigureDbDefault(v, "db-audit-ro", "CT_BRIDGE_DB_AUDIT_RO_USERNAME", "CT_BRIDGE_DB_AUDIT_RO_PASSWORD")

	//Storage custom configuration in DB (read/write)
	v.SetDefault("config-db-rw", true)
	database.ConfigureDbDefault(v, "db-config-rw", "CT_BRIDGE_DB_CONFIG_RW_USERNAME", "CT_BRIDGE_DB_CONFIG_RW_PASSWORD")

	v.SetDefault("db-config-rw-migration", false)
	v.SetDefault("db-config-rw-migration-version", "")

	//Storage custom configuration in DB (read only)
	v.SetDefault("config-db-ro", true)
	database.ConfigureDbDefault(v, "db-config-ro", "CT_BRIDGE_DB_CONFIG_RO_USERNAME", "CT_BRIDGE_DB_CONFIG_RO_PASSWORD")

	v.SetDefault("db-config-ro-migration", false)
	v.SetDefault("db-config-ro-migration-version", "")

	// Rate limiting (in requests/second)
	v.SetDefault("rate-event", 1000)
	v.SetDefault("rate-account", 1000)
	v.SetDefault("rate-management", 1000)
	v.SetDefault("rate-statistics", 1000)
	v.SetDefault("rate-events", 1000)

	// Influx DB client default.
	v.SetDefault("influx", false)
	v.SetDefault("influx-host-port", "")
	v.SetDefault("influx-username", "")
	v.SetDefault("influx-password", "")
	v.SetDefault("influx-database", "")
	v.SetDefault("influx-precision", "")
	v.SetDefault("influx-retention-policy", "")
	v.SetDefault("influx-write-consistency", "")
	v.SetDefault("influx-write-interval", "1s")

	// Sentry client default.
	v.SetDefault("sentry", false)
	v.SetDefault("sentry-dsn", "")

	// Jaeger tracing default.
	v.SetDefault("jaeger", false)
	v.SetDefault("jaeger-sampler-type", "")
	v.SetDefault("jaeger-sampler-param", 0)
	v.SetDefault("jaeger-sampler-host-port", "")
	v.SetDefault("jaeger-reporter-logspan", false)
	v.SetDefault("jaeger-write-interval", "1s")

	// Debug routes enabled.
	v.SetDefault("pprof-route-enabled", true)

	// First level of override.
	pflag.String("config-file", v.GetString("config-file"), "The configuration file path can be relative or absolute.")
	pflag.String("authorization-file", v.GetString("authorization-file"), "The authorization file path can be relative or absolute.")
	v.BindPFlag("config-file", pflag.Lookup("config-file"))
	v.BindPFlag("authorization-file", pflag.Lookup("authorization-file"))
	pflag.Parse()

	// Bind ENV variables
	// We use env variables to bind Openshift secrets
	var censoredParameters = map[string]bool{}

	v.BindEnv("influx-username", "CT_BRIDGE_INFLUX_USERNAME")
	v.BindEnv("influx-password", "CT_BRIDGE_INFLUX_PASSWORD")
	censoredParameters["influx-password"] = true

	v.BindEnv("sentry-dsn", "CT_BRIDGE_SENTRY_DSN")
	censoredParameters["sentry-dsn"] = true

	v.BindEnv("event-basic-auth-token", "CT_BRIDGE_EVENT_BASIC_AUTH")
	censoredParameters["event-basic-auth-token"] = true

	// Load and log config.
	v.SetConfigFile(v.GetString("config-file"))
	var err = v.ReadInConfig()
	if err != nil {
		logger.Error("error", err)
	}

	// If the host/port is not set, we consider the components deactivated.
	v.Set("influx", v.GetString("influx-host-port") != "")
	v.Set("sentry", v.GetString("sentry-dsn") != "")
	v.Set("jaeger", v.GetString("jaeger-sampler-host-port") != "")

	// Log config in alphabetical order.
	var keys = v.AllKeys()
	sort.Strings(keys)

	for _, k := range keys {
		if _, censored := censoredParameters[k]; censored {
			logger.Info(k, "*************")
		} else if strings.Contains(k, "password") {
			logger.Info(k, "*************")
		} else {
			logger.Info(k, v.Get(k))
		}
	}
	return v
}

func configureEventsHandler(ComponentName string, ComponentID string, idGenerator idgenerator.IDGenerator, keycloakClient *keycloak.Client, audienceRequired string, tracer tracing.OpentracingClient, logger log.Logger) func(endpoint endpoint.Endpoint) http.Handler {
	return func(endpoint endpoint.Endpoint) http.Handler {
		var handler http.Handler
		handler = events.MakeEventsHandler(endpoint, logger)
		handler = middleware.MakeHTTPCorrelationIDMW(idGenerator, tracer, logger, ComponentName, ComponentID)(handler)
		handler = middleware.MakeHTTPOIDCTokenValidationMW(keycloakClient, audienceRequired, logger)(handler)
		return handler
	}
}

func configureStatisiticsHandler(ComponentName string, ComponentID string, idGenerator idgenerator.IDGenerator, keycloakClient *keycloak.Client, audienceRequired string, tracer tracing.OpentracingClient, logger log.Logger) func(endpoint endpoint.Endpoint) http.Handler {
	return func(endpoint endpoint.Endpoint) http.Handler {
		var handler http.Handler
		handler = statistics.MakeStatisticsHandler(endpoint, logger)
		handler = middleware.MakeHTTPCorrelationIDMW(idGenerator, tracer, logger, ComponentName, ComponentID)(handler)
		handler = middleware.MakeHTTPOIDCTokenValidationMW(keycloakClient, audienceRequired, logger)(handler)
		return handler
	}
}

func configureManagementHandler(ComponentName string, ComponentID string, idGenerator idgenerator.IDGenerator, keycloakClient *keycloak.Client, audienceRequired string, tracer tracing.OpentracingClient, logger log.Logger) func(endpoint endpoint.Endpoint) http.Handler {
	return func(endpoint endpoint.Endpoint) http.Handler {
		var handler http.Handler
		handler = management.MakeManagementHandler(endpoint, logger)
		handler = middleware.MakeHTTPCorrelationIDMW(idGenerator, tracer, logger, ComponentName, ComponentID)(handler)
		handler = middleware.MakeHTTPOIDCTokenValidationMW(keycloakClient, audienceRequired, logger)(handler)
		return handler
	}
}

func configureRightsHandler(ComponentName string, ComponentID string, idGenerator idgenerator.IDGenerator, authorizationManager security.AuthorizationManager, keycloakClient *keycloak.Client, audienceRequired string, tracer tracing.OpentracingClient, logger log.Logger) http.Handler {
	var handler http.Handler
	handler = commonhttp.MakeRightsHandler(authorizationManager)
	handler = middleware.MakeHTTPCorrelationIDMW(idGenerator, tracer, logger, ComponentName, ComponentID)(handler)
	handler = middleware.MakeHTTPOIDCTokenValidationMW(keycloakClient, audienceRequired, logger)(handler)
	return handler
}

func configureAccountHandler(ComponentName string, ComponentID string, idGenerator idgenerator.IDGenerator, keycloakClient *keycloak.Client, audienceRequired string, tracer tracing.OpentracingClient, logger log.Logger) func(endpoint endpoint.Endpoint) http.Handler {
	return func(endpoint endpoint.Endpoint) http.Handler {
		var handler http.Handler
		handler = account.MakeAccountHandler(endpoint, logger)
		handler = middleware.MakeHTTPCorrelationIDMW(idGenerator, tracer, logger, ComponentName, ComponentID)(handler)
		handler = middleware.MakeHTTPOIDCTokenValidationMW(keycloakClient, audienceRequired, logger)(handler)
		return handler
	}
}

func configureEventsDbModule(baseEventsDBModule database.EventsDBModule, influxMetrics metrics.Metrics, logger log.Logger, tracer tracing.OpentracingClient) database.EventsDBModule {
	eventsDBModule := event.MakeEventsDBModuleInstrumentingMW(influxMetrics.NewHistogram("eventsDB_module"))(baseEventsDBModule)
	eventsDBModule = event.MakeEventsDBModuleLoggingMW(log.With(logger, "mw", "module", "unit", "eventsDB"))(eventsDBModule)
	eventsDBModule = event.MakeEventsDBModuleTracingMW(tracer)(eventsDBModule)
	return eventsDBModule
}

func prepareEndpoint(e cs.Endpoint, endpointName string, influxMetrics metrics.Metrics, managementLogger log.Logger, tracer tracing.OpentracingClient, rateLimit int) endpoint.Endpoint {
	e = middleware.MakeEndpointInstrumentingMW(influxMetrics, endpointName)(e)
	e = middleware.MakeEndpointLoggingMW(log.With(managementLogger, "mw", endpointName))(e)
	e = tracer.MakeEndpointTracingMW(endpointName)(e)
	return keycloakb.LimitRate(e, rateLimit)
}
