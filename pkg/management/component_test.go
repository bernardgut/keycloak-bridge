package management

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	cs "github.com/cloudtrust/common-service"
	"github.com/cloudtrust/common-service/database"
	commonhttp "github.com/cloudtrust/common-service/errors"
	"github.com/cloudtrust/common-service/log"
	api "github.com/cloudtrust/keycloak-bridge/api/management"
	"github.com/cloudtrust/keycloak-bridge/internal/dto"
	"github.com/cloudtrust/keycloak-bridge/internal/keycloakb"
	"github.com/cloudtrust/keycloak-bridge/pkg/management/mock"
	kc "github.com/cloudtrust/keycloak-client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetRealms(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="

	// Get realms with succces
	{
		var id = "1245"
		var keycloakVersion = "4.8.3"
		var realm = "master"
		var displayName = "Master"
		var enabled = true

		var kcRealmRep = kc.RealmRepresentation{
			Id:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		var kcRealmsRep []kc.RealmRepresentation
		kcRealmsRep = append(kcRealmsRep, kcRealmRep)

		mockKeycloakClient.EXPECT().GetRealms(accessToken).Return(kcRealmsRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiRealmsRep, err := managementComponent.GetRealms(ctx)

		var expectedAPIRealmRep = api.RealmRepresentation{
			ID:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		var expectedAPIRealmsRep []api.RealmRepresentation
		expectedAPIRealmsRep = append(expectedAPIRealmsRep, expectedAPIRealmRep)

		assert.Nil(t, err)
		assert.Equal(t, expectedAPIRealmsRep, apiRealmsRep)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetRealms(accessToken).Return([]kc.RealmRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetRealms(ctx)

		assert.NotNil(t, err)
	}
}

func TestGetRealm(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var username = "username"

	// Get realm with succces
	{
		var id = "1245"
		var keycloakVersion = "4.8.3"
		var realm = "master"
		var displayName = "Master"
		var enabled = true

		var kcRealmRep = kc.RealmRepresentation{
			Id:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmName).Return(kcRealmRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().Store(ctx, gomock.Any()).Return(nil).AnyTimes()

		apiRealmRep, err := managementComponent.GetRealm(ctx, "master")

		var expectedAPIRealmRep = api.RealmRepresentation{
			ID:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		assert.Nil(t, err)
		assert.Equal(t, expectedAPIRealmRep, apiRealmRep)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmName).Return(kc.RealmRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		//mockEventDBModule.EXPECT().Store(ctx, gomock.Any()).Return(nil).Times(1)

		_, err := managementComponent.GetRealm(ctx, "master")

		assert.NotNil(t, err)
	}
}

func TestGetClient(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"

	// Get client with succces
	{
		var id = "1245-1245-4578"
		var name = "clientName"
		var baseURL = "http://toto.com"
		var clientID = "client-id"
		var protocol = "saml"
		var enabled = true
		var username = "username"

		var kcClientRep = kc.ClientRepresentation{
			Id:       &id,
			Name:     &name,
			BaseUrl:  &baseURL,
			ClientId: &clientID,
			Protocol: &protocol,
			Enabled:  &enabled,
		}

		mockKeycloakClient.EXPECT().GetClient(accessToken, realmName, id).Return(kcClientRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().Store(ctx, gomock.Any()).Return(nil).AnyTimes()

		apiClientRep, err := managementComponent.GetClient(ctx, "master", id)

		var expectedAPIClientRep = api.ClientRepresentation{
			ID:       &id,
			Name:     &name,
			BaseURL:  &baseURL,
			ClientID: &clientID,
			Protocol: &protocol,
			Enabled:  &enabled,
		}

		assert.Nil(t, err)
		assert.Equal(t, expectedAPIClientRep, apiClientRep)
	}

	//Error
	{
		var id = "1234-79894-7594"
		mockKeycloakClient.EXPECT().GetClient(accessToken, realmName, id).Return(kc.ClientRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetClient(ctx, "master", id)

		assert.NotNil(t, err)
	}
}

func TestGetClients(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"

	// Get clients with succces
	{
		var id = "1234-7894-58"
		var name = "clientName"
		var baseURL = "http://toto.com"
		var clientID = "client-id"
		var protocol = "saml"
		var enabled = true

		var kcClientRep = kc.ClientRepresentation{
			Id:       &id,
			Name:     &name,
			BaseUrl:  &baseURL,
			ClientId: &clientID,
			Protocol: &protocol,
			Enabled:  &enabled,
		}

		var kcClientsRep []kc.ClientRepresentation
		kcClientsRep = append(kcClientsRep, kcClientRep)

		mockKeycloakClient.EXPECT().GetClients(accessToken, realmName).Return(kcClientsRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiClientsRep, err := managementComponent.GetClients(ctx, "master")

		var expectedAPIClientRep = api.ClientRepresentation{
			ID:       &id,
			Name:     &name,
			BaseURL:  &baseURL,
			ClientID: &clientID,
			Protocol: &protocol,
			Enabled:  &enabled,
		}

		var expectedAPIClientsRep []api.ClientRepresentation
		expectedAPIClientsRep = append(expectedAPIClientsRep, expectedAPIClientRep)

		assert.Nil(t, err)
		assert.Equal(t, expectedAPIClientsRep, apiClientsRep)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetClients(accessToken, realmName).Return([]kc.ClientRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetClients(ctx, "master")

		assert.NotNil(t, err)
	}
}

func TestCreateUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var username = "test"
	var realmName = "master"
	var targetRealmName = "DEP"
	var userID = "41dbf4a8-32a9-4000-8c17-edc854c31231"
	var locationURL = "http://toto.com/realms/" + userID

	// Create with minimum properties
	{
		var kcUserRep = kc.UserRepresentation{
			Username: &username,
		}

		mockKeycloakClient.EXPECT().CreateUser(accessToken, realmName, targetRealmName, kcUserRep).Return(locationURL, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "API_ACCOUNT_CREATION", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		var userRep = api.UserRepresentation{
			Username: &username,
		}

		location, err := managementComponent.CreateUser(ctx, targetRealmName, userRep)

		assert.Nil(t, err)
		assert.Equal(t, locationURL, location)
	}
	//Create with minimum properties and having error when storing the event
	{
		var kcUserRep = kc.UserRepresentation{
			Username: &username,
		}

		mockKeycloakClient.EXPECT().CreateUser(accessToken, realmName, realmName, kcUserRep).Return(locationURL, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "API_ACCOUNT_CREATION", "back-office", database.CtEventRealmName, realmName, database.CtEventUserID, userID, database.CtEventUsername, username).Return(errors.New("error")).Times(1)
		m := map[string]interface{}{"event_name": "API_ACCOUNT_CREATION", database.CtEventRealmName: realmName, database.CtEventUserID: userID, database.CtEventUsername: username}
		eventJSON, _ := json.Marshal(m)
		mockLogger.EXPECT().Error("err", "error", "event", string(eventJSON))

		var userRep = api.UserRepresentation{
			Username: &username,
		}

		location, err := managementComponent.CreateUser(ctx, realmName, userRep)

		assert.Nil(t, err)
		assert.Equal(t, locationURL, location)

	}

	// Create with all properties allowed by Bridge API
	{
		var email = "toto@elca.ch"
		var enabled = true
		var emailVerified = true
		var firstName = "Titi"
		var lastName = "Tutu"
		var phoneNumber = "+41789456"
		var phoneNumberVerified = true
		var label = "Label"
		var gender = "M"
		var birthDate = "01/01/1988"
		var userID = "1234-7558-7645"
		var locale = "de"

		mockKeycloakClient.EXPECT().CreateUser(accessToken, realmName, targetRealmName, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, targetRealmName string, kcUserRep kc.UserRepresentation) (string, error) {
				assert.Equal(t, username, *kcUserRep.Username)
				assert.Equal(t, email, *kcUserRep.Email)
				assert.Equal(t, enabled, *kcUserRep.Enabled)
				assert.Equal(t, emailVerified, *kcUserRep.EmailVerified)
				assert.Equal(t, firstName, *kcUserRep.FirstName)
				assert.Equal(t, lastName, *kcUserRep.LastName)
				assert.Equal(t, phoneNumber, (*kcUserRep.Attributes)["phoneNumber"][0])
				verified, _ := strconv.ParseBool(((*kcUserRep.Attributes)["phoneNumberVerified"][0]))
				assert.Equal(t, phoneNumberVerified, verified)
				assert.Equal(t, label, (*kcUserRep.Attributes)["label"][0])
				assert.Equal(t, gender, (*kcUserRep.Attributes)["gender"][0])
				assert.Equal(t, birthDate, (*kcUserRep.Attributes)["birthDate"][0])
				assert.Equal(t, locale, (*kcUserRep.Attributes)["locale"][0])

				return locationURL, nil
			}).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().Store(ctx, gomock.Any()).Return(nil).AnyTimes()

		var userRep = api.UserRepresentation{
			ID:                  &userID,
			Username:            &username,
			Email:               &email,
			Enabled:             &enabled,
			EmailVerified:       &emailVerified,
			FirstName:           &firstName,
			LastName:            &lastName,
			PhoneNumber:         &phoneNumber,
			PhoneNumberVerified: &phoneNumberVerified,
			Label:               &label,
			Gender:              &gender,
			BirthDate:           &birthDate,
			Locale:              &locale,
		}
		mockEventDBModule.EXPECT().ReportEvent(ctx, "API_ACCOUNT_CREATION", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		location, err := managementComponent.CreateUser(ctx, targetRealmName, userRep)

		assert.Nil(t, err)
		assert.Equal(t, locationURL, location)
	}

	// Error from KC client
	{
		var kcUserRep = kc.UserRepresentation{}

		mockKeycloakClient.EXPECT().CreateUser(accessToken, realmName, targetRealmName, kcUserRep).Return("", fmt.Errorf("Invalid input")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)

		var userRep = api.UserRepresentation{}
		mockLogger.EXPECT().Warn("err", "Invalid input")

		location, err := managementComponent.CreateUser(ctx, targetRealmName, userRep)

		assert.NotNil(t, err)
		assert.Equal(t, "", location)
	}
}

func TestDeleteUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var userID = "41dbf4a8-32a9-4000-8c17-edc854c31231"
	var realmName = "master"
	var username = "username"

	// Delete user with success
	{
		mockKeycloakClient.EXPECT().DeleteUser(accessToken, realmName, userID).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "API_ACCOUNT_DELETION", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		err := managementComponent.DeleteUser(ctx, "master", userID)

		assert.Nil(t, err)
	}
	// Delete user with success but the having an error when storing the event in the DB
	{
		mockKeycloakClient.EXPECT().DeleteUser(accessToken, realmName, userID).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "API_ACCOUNT_DELETION", "back-office", database.CtEventRealmName, realmName, database.CtEventUserID, userID).Return(errors.New("error")).Times(1)
		m := map[string]interface{}{"event_name": "API_ACCOUNT_DELETION", database.CtEventRealmName: realmName, database.CtEventUserID: userID}
		eventJSON, _ := json.Marshal(m)
		mockLogger.EXPECT().Error("err", "error", "event", string(eventJSON))
		err := managementComponent.DeleteUser(ctx, "master", userID)

		assert.Nil(t, err)

	}

	// Error from KC client
	{
		mockKeycloakClient.EXPECT().DeleteUser(accessToken, realmName, userID).Return(fmt.Errorf("Invalid input")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockLogger.EXPECT().Warn("err", "Invalid input")

		err := managementComponent.DeleteUser(ctx, "master", userID)

		assert.NotNil(t, err)
	}
}

func TestGetUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var id = "41dbf4a8-32a9-4000-8c17-edc854c31231"
	var username = "username"
	// Get user with succces
	{

		var email = "toto@elca.ch"
		var enabled = true
		var emailVerified = true
		var firstName = "Titi"
		var lastName = "Tutu"
		var phoneNumber = "+41789456"
		var phoneNumberVerified = true
		var label = "Label"
		var gender = "M"
		var birthDate = "01/01/1988"
		var createdTimestamp = time.Now().UTC().Unix()
		var locale = "it"

		var attributes = make(map[string][]string)
		attributes["phoneNumber"] = []string{phoneNumber}
		attributes["label"] = []string{label}
		attributes["gender"] = []string{gender}
		attributes["birthDate"] = []string{birthDate}
		attributes["phoneNumberVerified"] = []string{strconv.FormatBool(phoneNumberVerified)}
		attributes["locale"] = []string{locale}

		var kcUserRep = kc.UserRepresentation{
			Id:               &id,
			Username:         &username,
			Email:            &email,
			Enabled:          &enabled,
			EmailVerified:    &emailVerified,
			FirstName:        &firstName,
			LastName:         &lastName,
			Attributes:       &attributes,
			CreatedTimestamp: &createdTimestamp,
		}

		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kcUserRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "GET_DETAILS", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		apiUserRep, err := managementComponent.GetUser(ctx, "master", id)

		assert.Nil(t, err)
		assert.Equal(t, username, *apiUserRep.Username)
		assert.Equal(t, email, *apiUserRep.Email)
		assert.Equal(t, enabled, *apiUserRep.Enabled)
		assert.Equal(t, emailVerified, *apiUserRep.EmailVerified)
		assert.Equal(t, firstName, *apiUserRep.FirstName)
		assert.Equal(t, lastName, *apiUserRep.LastName)
		assert.Equal(t, phoneNumber, *apiUserRep.PhoneNumber)
		assert.Equal(t, phoneNumberVerified, *apiUserRep.PhoneNumberVerified)
		assert.Equal(t, label, *apiUserRep.Label)
		assert.Equal(t, gender, *apiUserRep.Gender)
		assert.Equal(t, birthDate, *apiUserRep.BirthDate)
		assert.Equal(t, createdTimestamp, *apiUserRep.CreatedTimestamp)
		assert.Equal(t, locale, *apiUserRep.Locale)
	}
	//Get user with succces but with error when storing the event in the DB
	{

		var kcUserRep = kc.UserRepresentation{
			Id:       &id,
			Username: &username,
		}
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kcUserRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "GET_DETAILS", "back-office", database.CtEventRealmName, realmName, database.CtEventUserID, id, database.CtEventUsername, username).Return(errors.New("error")).Times(1)
		m := map[string]interface{}{"event_name": "GET_DETAILS", database.CtEventRealmName: realmName, database.CtEventUserID: id, database.CtEventUsername: username}
		eventJSON, _ := json.Marshal(m)
		mockLogger.EXPECT().Error("err", "error", "event", string(eventJSON))

		apiUserRep, err := managementComponent.GetUser(ctx, "master", id)
		assert.Nil(t, err)
		assert.Equal(t, username, *apiUserRep.Username)
	}

	//Error
	{
		var id = "1234-79894-7594"
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kc.UserRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockLogger.EXPECT().Warn("err", "Unexpected error")

		_, err := managementComponent.GetUser(ctx, "master", id)

		assert.NotNil(t, err)
	}
}

func TestUpdateUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var id = "41dbf4a8-32a9-4000-8c17-edc854c31231"
	var username = "username"
	var enabled = true
	// Update user with succces
	{

		var email = "toto@elca.ch"
		var emailVerified = true
		var firstName = "Titi"
		var lastName = "Tutu"
		var phoneNumber = "+41789456"
		var phoneNumberVerified = true
		var label = "Label"
		var gender = "M"
		var birthDate = "01/01/1988"
		var locale = "de"
		var createdTimestamp = time.Now().UTC().Unix()

		var attributes = make(map[string][]string)
		attributes["phoneNumber"] = []string{phoneNumber}
		attributes["label"] = []string{label}
		attributes["gender"] = []string{gender}
		attributes["birthDate"] = []string{birthDate}
		attributes["phoneNumberVerified"] = []string{strconv.FormatBool(phoneNumberVerified)}
		attributes["locale"] = []string{locale}

		var kcUserRep = kc.UserRepresentation{
			Id:               &id,
			Username:         &username,
			Email:            &email,
			Enabled:          &enabled,
			EmailVerified:    &emailVerified,
			FirstName:        &firstName,
			LastName:         &lastName,
			Attributes:       &attributes,
			CreatedTimestamp: &createdTimestamp,
		}

		var userRep = api.UserRepresentation{
			Username:            &username,
			Email:               &email,
			Enabled:             &enabled,
			EmailVerified:       &emailVerified,
			FirstName:           &firstName,
			LastName:            &lastName,
			PhoneNumber:         &phoneNumber,
			PhoneNumberVerified: &phoneNumberVerified,
			Label:               &label,
			Gender:              &gender,
			BirthDate:           &birthDate,
			Locale:              &locale,
		}

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "LOCK_ACCOUNT", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		mockEventDBModule.EXPECT().ReportEvent(ctx, "UNLOCK_ACCOUNT", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kcUserRep, nil).Times(1)

		mockKeycloakClient.EXPECT().UpdateUser(accessToken, realmName, id, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, id string, kcUserRep kc.UserRepresentation) error {
				assert.Equal(t, username, *kcUserRep.Username)
				assert.Equal(t, email, *kcUserRep.Email)
				assert.Equal(t, enabled, *kcUserRep.Enabled)
				assert.Equal(t, emailVerified, *kcUserRep.EmailVerified)
				assert.Equal(t, firstName, *kcUserRep.FirstName)
				assert.Equal(t, lastName, *kcUserRep.LastName)
				assert.Equal(t, phoneNumber, (*kcUserRep.Attributes)["phoneNumber"][0])
				verified, _ := strconv.ParseBool(((*kcUserRep.Attributes)["phoneNumberVerified"][0]))
				assert.Equal(t, phoneNumberVerified, verified)
				assert.Equal(t, label, (*kcUserRep.Attributes)["label"][0])
				assert.Equal(t, gender, (*kcUserRep.Attributes)["gender"][0])
				assert.Equal(t, birthDate, (*kcUserRep.Attributes)["birthDate"][0])
				assert.Equal(t, locale, (*kcUserRep.Attributes)["locale"][0])
				return nil
			}).Times(1)

		err := managementComponent.UpdateUser(ctx, "master", id, userRep)

		assert.Nil(t, err)

		//update by locking the user
		enabled = false
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, realmName, id, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, id string, kcUserRep kc.UserRepresentation) error {
				assert.Equal(t, username, *kcUserRep.Username)
				assert.Equal(t, email, *kcUserRep.Email)
				assert.Equal(t, enabled, *kcUserRep.Enabled)
				assert.Equal(t, emailVerified, *kcUserRep.EmailVerified)
				assert.Equal(t, firstName, *kcUserRep.FirstName)
				assert.Equal(t, lastName, *kcUserRep.LastName)
				assert.Equal(t, phoneNumber, (*kcUserRep.Attributes)["phoneNumber"][0])
				verified, _ := strconv.ParseBool(((*kcUserRep.Attributes)["phoneNumberVerified"][0]))
				assert.Equal(t, phoneNumberVerified, verified)
				assert.Equal(t, label, (*kcUserRep.Attributes)["label"][0])
				assert.Equal(t, gender, (*kcUserRep.Attributes)["gender"][0])
				assert.Equal(t, birthDate, (*kcUserRep.Attributes)["birthDate"][0])
				assert.Equal(t, locale, (*kcUserRep.Attributes)["locale"][0])
				return nil
			}).Times(1)

		var userRepLocked = api.UserRepresentation{
			Username:            &username,
			Email:               &email,
			Enabled:             &enabled,
			EmailVerified:       &emailVerified,
			FirstName:           &firstName,
			LastName:            &lastName,
			PhoneNumber:         &phoneNumber,
			PhoneNumberVerified: &phoneNumberVerified,
			Label:               &label,
			Gender:              &gender,
			BirthDate:           &birthDate,
			Locale:              &locale,
		}
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kcUserRep, nil).Times(1)

		err = managementComponent.UpdateUser(ctx, "master", id, userRepLocked)

		assert.Nil(t, err)
		// update by changing the email address
		var oldEmail = "toti@elca.ch"
		var oldkcUserRep = kc.UserRepresentation{
			Id:            &id,
			Email:         &oldEmail,
			EmailVerified: &emailVerified,
		}
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(oldkcUserRep, nil).Times(1)
		mockEventDBModule.EXPECT().ReportEvent(ctx, "LOCK_ACCOUNT", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, realmName, id, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, id string, kcUserRep kc.UserRepresentation) error {
				assert.Equal(t, email, *kcUserRep.Email)
				assert.Equal(t, false, *kcUserRep.EmailVerified)
				return nil
			}).Times(1)

		err = managementComponent.UpdateUser(ctx, "master", id, userRep)

		assert.Nil(t, err)

		// update by changing the phone number

		var oldNumber = "+41789467"
		var oldAttributes = make(map[string][]string)
		oldAttributes["phoneNumber"] = []string{oldNumber}
		oldAttributes["phoneNumberVerified"] = []string{strconv.FormatBool(phoneNumberVerified)}
		var oldkcUserRep2 = kc.UserRepresentation{
			Id:         &id,
			Attributes: &oldAttributes,
		}
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(oldkcUserRep2, nil).Times(1)
		mockEventDBModule.EXPECT().ReportEvent(ctx, "LOCK_ACCOUNT", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, realmName, id, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, id string, kcUserRep kc.UserRepresentation) error {
				verified, _ := strconv.ParseBool(((*kcUserRep.Attributes)["phoneNumberVerified"][0]))
				assert.Equal(t, phoneNumber, (*kcUserRep.Attributes)["phoneNumber"][0])
				assert.Equal(t, false, verified)
				return nil
			}).Times(1)

		err = managementComponent.UpdateUser(ctx, "master", id, userRep)

		assert.Nil(t, err)

		// update without attributes
		var userRepWithoutAttr = api.UserRepresentation{
			Username:  &username,
			Email:     &email,
			Enabled:   &enabled,
			FirstName: &firstName,
			LastName:  &lastName,
		}

		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(oldkcUserRep2, nil).Times(1)
		mockEventDBModule.EXPECT().ReportEvent(ctx, "LOCK_ACCOUNT", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, realmName, id, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, id string, kcUserRep kc.UserRepresentation) error {
				verified, _ := strconv.ParseBool(((*kcUserRep.Attributes)["phoneNumberVerified"][0]))
				assert.Equal(t, oldNumber, (*kcUserRep.Attributes)["phoneNumber"][0])
				assert.Equal(t, true, verified)
				return nil
			}).Times(1)

		err = managementComponent.UpdateUser(ctx, "master", id, userRepWithoutAttr)

		assert.Nil(t, err)
	}
	// Update user with succces but with error when storing the event in the DB
	{
		enabled = true
		var kcUserRep = kc.UserRepresentation{
			Id:       &id,
			Username: &username,
			Enabled:  &enabled,
		}

		var userRep = api.UserRepresentation{
			Username: &username,
			Enabled:  &enabled,
		}

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "UNLOCK_ACCOUNT", "back-office", database.CtEventRealmName, realmName, database.CtEventUserID, id, database.CtEventUsername, username).Return(errors.New("error")).Times(1)
		m := map[string]interface{}{"event_name": "UNLOCK_ACCOUNT", database.CtEventRealmName: realmName, database.CtEventUserID: id, database.CtEventUsername: username}
		eventJSON, _ := json.Marshal(m)
		mockLogger.EXPECT().Error("err", "error", "event", string(eventJSON))
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kcUserRep, nil).Times(1)
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, realmName, id, gomock.Any()).Return(nil).Times(1)

		err := managementComponent.UpdateUser(ctx, "master", id, userRep)

		assert.Nil(t, err)

	}

	//Error - get user
	{
		var id = "1234-79894-7594"
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kc.UserRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockLogger.EXPECT().Warn("err", "Unexpected error")
		err := managementComponent.UpdateUser(ctx, "master", id, api.UserRepresentation{})

		assert.NotNil(t, err)
	}
	//Error - update user
	{
		var id = "1234-79894-7594"
		var kcUserRep = kc.UserRepresentation{
			Id: &id,
		}
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, id).Return(kcUserRep, nil).AnyTimes()
		mockKeycloakClient.EXPECT().UpdateUser(accessToken, realmName, id, gomock.Any()).Return(fmt.Errorf("Unexpected error")).Times(1)
		mockLogger.EXPECT().Warn("err", "Unexpected error")
		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		err := managementComponent.UpdateUser(ctx, "master", id, api.UserRepresentation{})

		assert.NotNil(t, err)
	}
}

func TestGetUsers(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var targetRealmName = "DEP"

	// Get user with succces
	{
		var id = "1234-7454-4516"
		var username = "username"
		var email = "toto@elca.ch"
		var enabled = true
		var emailVerified = true
		var firstName = "Titi"
		var lastName = "Tutu"
		var phoneNumber = "+41789456"
		var phoneNumberVerified = true
		var label = "Label"
		var gender = "M"
		var birthDate = "01/01/1988"
		var createdTimestamp = time.Now().UTC().Unix()

		var attributes = make(map[string][]string)
		attributes["phoneNumber"] = []string{phoneNumber}
		attributes["label"] = []string{label}
		attributes["gender"] = []string{gender}
		attributes["birthDate"] = []string{birthDate}
		attributes["phoneNumberVerified"] = []string{strconv.FormatBool(phoneNumberVerified)}

		var count = 1
		var kcUserRep = kc.UserRepresentation{
			Id:               &id,
			Username:         &username,
			Email:            &email,
			Enabled:          &enabled,
			EmailVerified:    &emailVerified,
			FirstName:        &firstName,
			LastName:         &lastName,
			Attributes:       &attributes,
			CreatedTimestamp: &createdTimestamp,
		}
		var kcUsersRep = kc.UsersPageRepresentation{
			Count: &count,
			Users: []kc.UserRepresentation{kcUserRep},
		}

		mockKeycloakClient.EXPECT().GetUsers(accessToken, realmName, targetRealmName, "groupId", "123-456-789").Return(kcUsersRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, "master")

		apiUsersRep, err := managementComponent.GetUsers(ctx, "DEP", []string{"123-456-789"})

		var apiUserRep = apiUsersRep.Users[0]
		assert.Nil(t, err)
		assert.Equal(t, username, *apiUserRep.Username)
		assert.Equal(t, email, *apiUserRep.Email)
		assert.Equal(t, enabled, *apiUserRep.Enabled)
		assert.Equal(t, emailVerified, *apiUserRep.EmailVerified)
		assert.Equal(t, firstName, *apiUserRep.FirstName)
		assert.Equal(t, lastName, *apiUserRep.LastName)
		assert.Equal(t, phoneNumber, *apiUserRep.PhoneNumber)
		verified, _ := strconv.ParseBool(((*kcUserRep.Attributes)["phoneNumberVerified"][0]))
		assert.Equal(t, phoneNumberVerified, verified)
		assert.Equal(t, label, *apiUserRep.Label)
		assert.Equal(t, gender, *apiUserRep.Gender)
		assert.Equal(t, birthDate, *apiUserRep.BirthDate)
		assert.Equal(t, createdTimestamp, *apiUserRep.CreatedTimestamp)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetUsers(accessToken, realmName, targetRealmName, "groupId", "123-456-789").Return(kc.UsersPageRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, "master")

		_, err := managementComponent.GetUsers(ctx, "DEP", []string{"123-456-789"})

		assert.NotNil(t, err)
	}
}

func TestGetUserAccountStatus(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmReq = "master"
	var realmName = "aRealm"
	var userID = "789-789-456"

	// GetUser returns an error
	{
		var userRep kc.UserRepresentation
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, userID).Return(userRep, fmt.Errorf("Unexpected error")).Times(1)
		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		_, err := managementComponent.GetUserAccountStatus(ctx, realmName, userID)
		assert.NotNil(t, err)
	}

	// GetUser returns a non-enabled user
	{
		var userRep kc.UserRepresentation
		enabled := false
		userRep.Enabled = &enabled
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, userID).Return(userRep, nil).Times(1)
		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		status, err := managementComponent.GetUserAccountStatus(ctx, realmName, userID)
		assert.Nil(t, err)
		assert.False(t, status["enabled"])
	}

	// GetUser returns an enabled user but GetCredentialsForUser fails
	{
		var userRep kc.UserRepresentation
		enabled := true
		userRep.Enabled = &enabled
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, userID).Return(userRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return(nil, fmt.Errorf("Unexpected error")).Times(1)
		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)
		_, err := managementComponent.GetUserAccountStatus(ctx, realmName, userID)
		assert.NotNil(t, err)
	}

	// GetUser returns an enabled user but GetCredentialsForUser have no credential
	{
		var userRep kc.UserRepresentation
		enabled := true
		userRep.Enabled = &enabled
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, userID).Return(userRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return([]kc.CredentialRepresentation{}, nil).Times(1)
		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)
		status, err := managementComponent.GetUserAccountStatus(ctx, realmName, userID)
		assert.Nil(t, err)
		assert.False(t, status["enabled"])
	}

	// GetUser returns an enabled user and GetCredentialsForUser have credentials
	{
		var userRep kc.UserRepresentation
		var creds1, creds2 kc.CredentialRepresentation
		enabled := true
		userRep.Enabled = &enabled
		mockKeycloakClient.EXPECT().GetUser(accessToken, realmName, userID).Return(userRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return([]kc.CredentialRepresentation{creds1, creds2}, nil).Times(1)
		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)
		status, err := managementComponent.GetUserAccountStatus(ctx, realmName, userID)
		assert.Nil(t, err)
		assert.True(t, status["enabled"])
	}
}

func TestGetClientRolesForUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "789-789-456"
	var clientID = "456-789-147"

	// Get role with succces
	{
		var id = "1234-7454-4516"
		var composite = false
		var containerID = "containerId"
		var description = "description role"
		var clientRole = true
		var name = "client name"

		var kcRoleRep = kc.RoleRepresentation{
			Id:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerId: &containerID,
			Description: &description,
		}

		var kcRolesRep []kc.RoleRepresentation
		kcRolesRep = append(kcRolesRep, kcRoleRep)

		mockKeycloakClient.EXPECT().GetClientRoleMappings(accessToken, realmName, userID, clientID).Return(kcRolesRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiRolesRep, err := managementComponent.GetClientRolesForUser(ctx, "master", userID, clientID)

		var apiRoleRep = apiRolesRep[0]
		assert.Nil(t, err)
		assert.Equal(t, id, *apiRoleRep.ID)
		assert.Equal(t, name, *apiRoleRep.Name)
		assert.Equal(t, clientRole, *apiRoleRep.ClientRole)
		assert.Equal(t, composite, *apiRoleRep.Composite)
		assert.Equal(t, containerID, *apiRoleRep.ContainerID)
		assert.Equal(t, description, *apiRoleRep.Description)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetClientRoleMappings(accessToken, realmName, userID, clientID).Return([]kc.RoleRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetClientRolesForUser(ctx, "master", userID, clientID)

		assert.NotNil(t, err)
	}
}

func TestAddClientRolesToUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "789-789-456"
	var clientID = "456-789-147"

	// Add role with succces
	{
		var id = "1234-7454-4516"
		var composite = false
		var containerID = "containerId"
		var description = "description role"
		var clientRole = true
		var name = "client name"

		var kcRoleRep = kc.RoleRepresentation{
			Id:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerId: &containerID,
			Description: &description,
		}

		var kcRolesRep []kc.RoleRepresentation
		kcRolesRep = append(kcRolesRep, kcRoleRep)

		mockKeycloakClient.EXPECT().AddClientRolesToUserRoleMapping(accessToken, realmName, userID, clientID, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, userID, clientID string, roles []kc.RoleRepresentation) error {
				var role = roles[0]
				assert.Equal(t, id, *role.Id)
				assert.Equal(t, name, *role.Name)
				assert.Equal(t, clientRole, *role.ClientRole)
				assert.Equal(t, composite, *role.Composite)
				assert.Equal(t, containerID, *role.ContainerId)
				assert.Equal(t, description, *role.Description)
				return nil
			}).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		var roleRep = api.RoleRepresentation{
			ID:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerID: &containerID,
			Description: &description,
		}
		var rolesRep []api.RoleRepresentation
		rolesRep = append(rolesRep, roleRep)

		err := managementComponent.AddClientRolesToUser(ctx, "master", userID, clientID, rolesRep)

		assert.Nil(t, err)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().AddClientRolesToUserRoleMapping(accessToken, realmName, userID, clientID, gomock.Any()).Return(fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		err := managementComponent.AddClientRolesToUser(ctx, "master", userID, clientID, []api.RoleRepresentation{})

		assert.NotNil(t, err)
	}
}

func TestGetRolesOfUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "789-789-456"

	// Get role with succces
	{
		var id = "1234-7454-4516"
		var composite = false
		var containerID = "containerId"
		var description = "description role"
		var clientRole = false
		var name = "client name"

		var kcRoleRep = kc.RoleRepresentation{
			Id:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerId: &containerID,
			Description: &description,
		}

		var kcRolesRep []kc.RoleRepresentation
		kcRolesRep = append(kcRolesRep, kcRoleRep)

		mockKeycloakClient.EXPECT().GetRealmLevelRoleMappings(accessToken, realmName, userID).Return(kcRolesRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiRolesRep, err := managementComponent.GetRolesOfUser(ctx, "master", userID)

		var apiRoleRep = apiRolesRep[0]
		assert.Nil(t, err)
		assert.Equal(t, id, *apiRoleRep.ID)
		assert.Equal(t, name, *apiRoleRep.Name)
		assert.Equal(t, clientRole, *apiRoleRep.ClientRole)
		assert.Equal(t, composite, *apiRoleRep.Composite)
		assert.Equal(t, containerID, *apiRoleRep.ContainerID)
		assert.Equal(t, description, *apiRoleRep.Description)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetRealmLevelRoleMappings(accessToken, realmName, userID).Return([]kc.RoleRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetRolesOfUser(ctx, "master", userID)

		assert.NotNil(t, err)
	}
}

func TestGetGroupsOfUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "789-789-456"

	// Get groups with succces
	{
		var id = "1234-7454-4516"
		var name = "client name"

		var kcGroupRep = kc.GroupRepresentation{
			Id:   &id,
			Name: &name,
		}

		var kcGroupsRep []kc.GroupRepresentation
		kcGroupsRep = append(kcGroupsRep, kcGroupRep)

		mockKeycloakClient.EXPECT().GetGroupsOfUser(accessToken, realmName, userID).Return(kcGroupsRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiGroupsRep, err := managementComponent.GetGroupsOfUser(ctx, "master", userID)

		var apiGroupRep = apiGroupsRep[0]
		assert.Nil(t, err)
		assert.Equal(t, id, *apiGroupRep.ID)
		assert.Equal(t, name, *apiGroupRep.Name)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetGroupsOfUser(accessToken, realmName, userID).Return([]kc.GroupRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetGroupsOfUser(ctx, "master", userID)

		assert.NotNil(t, err)
	}
}

func TestResetPassword(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "41dbf4a8-32a9-4000-8c17-edc854c31231"
	var password = "P@ssw0rd"
	var typePassword = "password"
	var username = "username"

	// Change password
	{
		var kcCredRep = kc.CredentialRepresentation{
			Type:  &typePassword,
			Value: &password,
		}

		mockKeycloakClient.EXPECT().ResetPassword(accessToken, realmName, userID, kcCredRep).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "INIT_PASSWORD", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		var passwordRep = api.PasswordRepresentation{
			Value: &password,
		}

		_, err := managementComponent.ResetPassword(ctx, "master", userID, passwordRep)

		assert.Nil(t, err)
	}
	// Change password but with error when storing the DB
	{
		var kcCredRep = kc.CredentialRepresentation{
			Type:  &typePassword,
			Value: &password,
		}

		mockKeycloakClient.EXPECT().ResetPassword(accessToken, realmName, userID, kcCredRep).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "INIT_PASSWORD", "back-office", database.CtEventRealmName, realmName, database.CtEventUserID, userID).Return(errors.New("error")).Times(1)
		m := map[string]interface{}{"event_name": "INIT_PASSWORD", database.CtEventRealmName: realmName, database.CtEventUserID: userID}
		eventJSON, _ := json.Marshal(m)
		mockLogger.EXPECT().Error("err", "error", "event", string(eventJSON))
		var passwordRep = api.PasswordRepresentation{
			Value: &password,
		}

		_, err := managementComponent.ResetPassword(ctx, "master", userID, passwordRep)

		assert.Nil(t, err)
	}

	// No password offered
	{
		var id = "master_id"
		var keycloakVersion = "4.8.3"
		var realm = "master"
		var displayName = "Master"
		var enabled = true

		var policy = "forceExpiredPasswordChange(365) and specialChars(1) and upperCase(1) and lowerCase(1) and length(4) and digits(1) and notUsername(undefined)"
		var kcRealmRep = kc.RealmRepresentation{
			Id:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
			PasswordPolicy:  &policy,
		}

		mockKeycloakClient.EXPECT().ResetPassword(accessToken, realmName, userID, gomock.Any()).Return(nil).Times(1)
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmName).Return(kcRealmRep, nil).AnyTimes()

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "INIT_PASSWORD", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		var passwordRep = api.PasswordRepresentation{
			Value: nil,
		}

		pwd, err := managementComponent.ResetPassword(ctx, "master", userID, passwordRep)

		assert.Nil(t, err)
		assert.NotNil(t, pwd)
	}

	// No password offered, no keycloak policy
	{
		var id = "master_id"

		var kcRealmRep = kc.RealmRepresentation{
			Id: &id,
		}

		mockKeycloakClient.EXPECT().ResetPassword(accessToken, realmName, userID, gomock.Any()).Return(nil).Times(1)
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmName).Return(kcRealmRep, nil).AnyTimes()

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmName)
		ctx = context.WithValue(ctx, cs.CtContextUsername, username)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "INIT_PASSWORD", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		var passwordRep = api.PasswordRepresentation{
			Value: nil,
		}

		pwd, err := managementComponent.ResetPassword(ctx, "master", userID, passwordRep)

		assert.Nil(t, err)
		assert.NotNil(t, pwd)
	}
	// Error
	{
		mockKeycloakClient.EXPECT().ResetPassword(accessToken, realmName, userID, gomock.Any()).Return(fmt.Errorf("Invalid input")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		var passwordRep = api.PasswordRepresentation{
			Value: &password,
		}
		mockLogger.EXPECT().Warn("err", "Invalid input")
		_, err := managementComponent.ResetPassword(ctx, "master", userID, passwordRep)

		assert.NotNil(t, err)
	}

}

func TestSendVerifyEmail(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "1245-7854-8963"

	var key1 = "key1"
	var value1 = "value1"
	var key2 = "key2"
	var value2 = "value2"

	// Send email
	{

		mockKeycloakClient.EXPECT().SendVerifyEmail(accessToken, realmName, userID, key1, value1, key2, value2).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		err := managementComponent.SendVerifyEmail(ctx, "master", userID, key1, value1, key2, value2)

		assert.Nil(t, err)
	}

	// Error
	{
		mockKeycloakClient.EXPECT().SendVerifyEmail(accessToken, realmName, userID).Return(fmt.Errorf("Invalid input")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		err := managementComponent.SendVerifyEmail(ctx, "master", userID)

		assert.NotNil(t, err)
	}
}

func TestExecuteActionsEmail(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "1245-7854-8963"
	var reqActions = []api.RequiredAction{initPasswordAction, "action1", "action2"}
	var actions = []string{initPasswordAction, "action1", "action2"}

	var key1 = "key1"
	var value1 = "value1"
	var key2 = "key2"
	var value2 = "value2"

	// Send email actions
	{

		mockKeycloakClient.EXPECT().ExecuteActionsEmail(accessToken, realmName, userID, actions, key1, value1, key2, value2).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockEventDBModule.EXPECT().ReportEvent(ctx, "INIT_PASSWORD", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)

		err := managementComponent.ExecuteActionsEmail(ctx, "master", userID, reqActions, key1, value1, key2, value2)

		assert.Nil(t, err)
	}

	// Error
	{
		mockKeycloakClient.EXPECT().ExecuteActionsEmail(accessToken, realmName, userID, actions).Return(fmt.Errorf("Invalid input")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		err := managementComponent.ExecuteActionsEmail(ctx, "master", userID, reqActions)

		assert.NotNil(t, err)
	}
}

func TestSendNewEnrolmentCode(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "1245-7854-8963"

	// Send new enrolment code
	{
		var code = "1234"
		mockKeycloakClient.EXPECT().SendNewEnrolmentCode(accessToken, realmName, userID).Return(kc.SmsCodeRepresentation{Code: &code}, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "SMS_CHALLENGE", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		codeRes, err := managementComponent.SendNewEnrolmentCode(ctx, "master", userID)

		assert.Nil(t, err)
		assert.Equal(t, "1234", codeRes)
	}
	// Send new enrolment code but have error when storing the event in the DB
	{
		var code = "1234"
		mockKeycloakClient.EXPECT().SendNewEnrolmentCode(accessToken, realmName, userID).Return(kc.SmsCodeRepresentation{Code: &code}, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		mockEventDBModule.EXPECT().ReportEvent(ctx, "SMS_CHALLENGE", "back-office", database.CtEventRealmName, realmName, database.CtEventUserID, userID).Return(errors.New("error")).Times(1)
		m := map[string]interface{}{"event_name": "SMS_CHALLENGE", database.CtEventRealmName: realmName, database.CtEventUserID: userID}
		eventJSON, _ := json.Marshal(m)
		mockLogger.EXPECT().Error("err", "error", "event", string(eventJSON))
		codeRes, err := managementComponent.SendNewEnrolmentCode(ctx, "master", userID)

		assert.Nil(t, err)
		assert.Equal(t, "1234", codeRes)
	}
	// Error
	{
		mockKeycloakClient.EXPECT().SendNewEnrolmentCode(accessToken, realmName, userID).Return(kc.SmsCodeRepresentation{}, fmt.Errorf("Invalid input")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockLogger.EXPECT().Warn("err", "Invalid input")
		_, err := managementComponent.SendNewEnrolmentCode(ctx, "master", userID)

		assert.NotNil(t, err)
	}
}

func TestSendReminderEmail(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var userID = "1245-7854-8963"

	var key1 = "key1"
	var value1 = "value1"
	var key2 = "key2"
	var value2 = "value2"
	var key3 = "key3"
	var value3 = "value3"

	// Send email
	{

		mockKeycloakClient.EXPECT().SendReminderEmail(accessToken, realmName, userID, key1, value1, key2, value2, key3, value3).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		err := managementComponent.SendReminderEmail(ctx, "master", userID, key1, value1, key2, value2, key3, value3)

		assert.Nil(t, err)
	}

	// Error
	{
		mockKeycloakClient.EXPECT().SendReminderEmail(accessToken, realmName, userID).Return(fmt.Errorf("Invalid input")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		err := managementComponent.SendReminderEmail(ctx, "master", userID)

		assert.NotNil(t, err)
	}
}

func TestGetCredentialsForUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)
	var accessToken = "TOKEN=="
	var realmReq = "master"
	var realmName = "otherRealm"
	var userID = "1245-7854-8963"

	// Get credentials for user
	{
		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return([]kc.CredentialRepresentation{}, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)

		_, err := managementComponent.GetCredentialsForUser(ctx, realmName, userID)

		assert.Nil(t, err)
	}
}

func TestDeleteCredentialsForUser(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)
	var accessToken = "TOKEN=="
	var realmReq = "master"
	var realmName = "master"
	var userID = "1245-7854-8963"
	var credential = "987-654-321"

	// Delete credentials for user
	{
		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return([]kc.CredentialRepresentation{}, nil).Times(1)
		mockKeycloakClient.EXPECT().DeleteCredential(accessToken, realmName, userID, credential).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)

		err := managementComponent.DeleteCredentialsForUser(ctx, realmName, userID, credential)

		assert.Nil(t, err)
	}
	// Delete credentials for user - error at obtaining the list of credentials
	{
		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return([]kc.CredentialRepresentation{}, errors.New("error")).Times(1)
		mockLogger.EXPECT().Warn("msg", "Could not obtain list of credentials", "err", "error")
		mockKeycloakClient.EXPECT().DeleteCredential(accessToken, realmName, userID, credential).Return(nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)

		err := managementComponent.DeleteCredentialsForUser(ctx, realmName, userID, credential)

		assert.Nil(t, err)

	}
	// Delete credentials for user - error at deleting the credential
	{
		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return([]kc.CredentialRepresentation{}, nil).Times(1)
		mockKeycloakClient.EXPECT().DeleteCredential(accessToken, realmName, userID, credential).Return(errors.New("error")).Times(1)
		mockLogger.EXPECT().Warn("err", "error")
		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)

		err := managementComponent.DeleteCredentialsForUser(ctx, realmName, userID, credential)

		assert.NotNil(t, err)

	}
	// Delete credentials for user
	{

		pwdId := "51389847-08f4-4a0f-9f9c-694554e626f2"
		pwd := "password"
		var credKcPwd = kc.CredentialRepresentation{
			Id:   &pwdId,
			Type: &pwd,
		}

		otpId := "51389847-08f4-4a0f-9f9c-694554e626f3"
		totp := "totp"
		var credKcOtp = kc.CredentialRepresentation{
			Id:   &otpId,
			Type: &totp,
		}

		var credsKc []kc.CredentialRepresentation
		credsKc = append(credsKc, credKcPwd)
		credsKc = append(credsKc, credKcOtp)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)

		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return(credsKc, nil).Times(1)
		mockKeycloakClient.EXPECT().DeleteCredential(accessToken, realmName, userID, otpId).Return(nil).Times(1)
		mockEventDBModule.EXPECT().ReportEvent(ctx, "2ND_FACTOR_REMOVED", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

		err := managementComponent.DeleteCredentialsForUser(ctx, realmName, userID, otpId)

		assert.Nil(t, err)
	}
	// Delete credentials for user - error at storing the event
	{

		pwdId := "51389847-08f4-4a0f-9f9c-694554e626f2"
		pwd := "password"
		var credKcPwd = kc.CredentialRepresentation{
			Id:   &pwdId,
			Type: &pwd,
		}

		otpId := "51389847-08f4-4a0f-9f9c-694554e626f3"
		totp := "totp"
		var credKcOtp = kc.CredentialRepresentation{
			Id:   &otpId,
			Type: &totp,
		}

		var credsKc []kc.CredentialRepresentation
		credsKc = append(credsKc, credKcPwd)
		credsKc = append(credsKc, credKcOtp)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		ctx = context.WithValue(ctx, cs.CtContextRealm, realmReq)

		mockKeycloakClient.EXPECT().GetCredentials(accessToken, realmName, userID).Return(credsKc, nil).Times(1)
		mockKeycloakClient.EXPECT().DeleteCredential(accessToken, realmName, userID, otpId).Return(nil).Times(1)
		mockEventDBModule.EXPECT().ReportEvent(ctx, "2ND_FACTOR_REMOVED", "back-office", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)
		m := map[string]interface{}{"event_name": "2ND_FACTOR_REMOVED", database.CtEventRealmName: realmName, database.CtEventUserID: userID}
		eventJSON, _ := json.Marshal(m)
		mockLogger.EXPECT().Error("err", "error", "event", string(eventJSON))
		err := managementComponent.DeleteCredentialsForUser(ctx, realmName, userID, otpId)

		assert.Nil(t, err)
	}

}

func TestGetRoles(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"

	// Get roles with succces
	{
		var id = "1234-7454-4516"
		var composite = false
		var containerID = "containerId"
		var description = "description role"
		var clientRole = false
		var name = "name"

		var kcRoleRep = kc.RoleRepresentation{
			Id:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerId: &containerID,
			Description: &description,
		}

		var kcRolesRep []kc.RoleRepresentation
		kcRolesRep = append(kcRolesRep, kcRoleRep)

		mockKeycloakClient.EXPECT().GetRoles(accessToken, realmName).Return(kcRolesRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiRolesRep, err := managementComponent.GetRoles(ctx, "master")

		var apiRoleRep = apiRolesRep[0]
		assert.Nil(t, err)
		assert.Equal(t, id, *apiRoleRep.ID)
		assert.Equal(t, name, *apiRoleRep.Name)
		assert.Equal(t, clientRole, *apiRoleRep.ClientRole)
		assert.Equal(t, composite, *apiRoleRep.Composite)
		assert.Equal(t, containerID, *apiRoleRep.ContainerID)
		assert.Equal(t, description, *apiRoleRep.Description)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetRoles(accessToken, realmName).Return([]kc.RoleRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetRoles(ctx, "master")

		assert.NotNil(t, err)
	}
}

func TestGetRole(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"

	// Get roles with succces
	{
		var id = "1234-7454-4516"
		var composite = false
		var containerID = "containerId"
		var description = "description role"
		var clientRole = false
		var name = "name"

		var kcRoleRep = kc.RoleRepresentation{
			Id:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerId: &containerID,
			Description: &description,
		}

		mockKeycloakClient.EXPECT().GetRole(accessToken, realmName, id).Return(kcRoleRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiRoleRep, err := managementComponent.GetRole(ctx, "master", id)

		assert.Nil(t, err)
		assert.Equal(t, id, *apiRoleRep.ID)
		assert.Equal(t, name, *apiRoleRep.Name)
		assert.Equal(t, clientRole, *apiRoleRep.ClientRole)
		assert.Equal(t, composite, *apiRoleRep.Composite)
		assert.Equal(t, containerID, *apiRoleRep.ContainerID)
		assert.Equal(t, description, *apiRoleRep.Description)
	}

	//Error
	{
		var id = "1234-7454-4516"
		mockKeycloakClient.EXPECT().GetRole(accessToken, realmName, id).Return(kc.RoleRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetRole(ctx, "master", id)

		assert.NotNil(t, err)
	}
}

func TestGetGroups(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"

	// Get groups with succces, non empty result
	{
		var id = "1234-7454-4516"
		var path = "path_group"
		var name = "group1"
		var realmRoles = []string{"role1"}

		var kcGroupRep = kc.GroupRepresentation{
			Id:         &id,
			Name:       &name,
			Path:       &path,
			RealmRoles: &realmRoles,
		}

		var kcGroupsRep []kc.GroupRepresentation
		kcGroupsRep = append(kcGroupsRep, kcGroupRep)

		mockKeycloakClient.EXPECT().GetGroups(accessToken, realmName).Return(kcGroupsRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiGroupsRep, err := managementComponent.GetGroups(ctx, "master")

		var apiGroupRep = apiGroupsRep[0]
		assert.Nil(t, err)
		assert.Equal(t, id, *apiGroupRep.ID)
		assert.Equal(t, name, *apiGroupRep.Name)
	}

	// Get groups with success, empty result
	{
		var kcGroupsRep []kc.GroupRepresentation
		mockKeycloakClient.EXPECT().GetGroups(accessToken, realmName).Return(kcGroupsRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		apiGroupsRep, err := managementComponent.GetGroups(ctx, "master")

		assert.Nil(t, err)
		assert.NotNil(t, apiGroupsRep)
		assert.Equal(t, 0, len(apiGroupsRep))
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetGroups(accessToken, realmName).Return([]kc.GroupRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetGroups(ctx, "master")

		assert.NotNil(t, err)
	}
}

func TestGetClientRoles(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var clientID = "15436-464-4"

	// Get roles with succces
	{
		var id = "1234-7454-4516"
		var composite = false
		var containerID = "containerId"
		var description = "description role"
		var clientRole = true
		var name = "name"

		var kcRoleRep = kc.RoleRepresentation{
			Id:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerId: &containerID,
			Description: &description,
		}

		var kcRolesRep []kc.RoleRepresentation
		kcRolesRep = append(kcRolesRep, kcRoleRep)

		mockKeycloakClient.EXPECT().GetClientRoles(accessToken, realmName, clientID).Return(kcRolesRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		apiRolesRep, err := managementComponent.GetClientRoles(ctx, "master", clientID)

		var apiRoleRep = apiRolesRep[0]
		assert.Nil(t, err)
		assert.Equal(t, id, *apiRoleRep.ID)
		assert.Equal(t, name, *apiRoleRep.Name)
		assert.Equal(t, clientRole, *apiRoleRep.ClientRole)
		assert.Equal(t, composite, *apiRoleRep.Composite)
		assert.Equal(t, containerID, *apiRoleRep.ContainerID)
		assert.Equal(t, description, *apiRoleRep.Description)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().GetClientRoles(accessToken, realmName, clientID).Return([]kc.RoleRepresentation{}, fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetClientRoles(ctx, "master", clientID)

		assert.NotNil(t, err)
	}
}

func TestCreateClientRole(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmName = "master"
	var clientID = "456-789-147"

	// Add role with succces
	{
		var id = "1234-7454-4516"
		var composite = false
		var containerID = "containerId"
		var description = "description role"
		var clientRole = true
		var name = "client name"

		var locationURL = "http://location.url"

		mockKeycloakClient.EXPECT().CreateClientRole(accessToken, realmName, clientID, gomock.Any()).DoAndReturn(
			func(accessToken, realmName, clientID string, role kc.RoleRepresentation) (string, error) {
				assert.Equal(t, id, *role.Id)
				assert.Equal(t, name, *role.Name)
				assert.Equal(t, clientRole, *role.ClientRole)
				assert.Equal(t, composite, *role.Composite)
				assert.Equal(t, containerID, *role.ContainerId)
				assert.Equal(t, description, *role.Description)
				return locationURL, nil
			}).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		var roleRep = api.RoleRepresentation{
			ID:          &id,
			Name:        &name,
			ClientRole:  &clientRole,
			Composite:   &composite,
			ContainerID: &containerID,
			Description: &description,
		}

		location, err := managementComponent.CreateClientRole(ctx, "master", clientID, roleRep)

		assert.Nil(t, err)
		assert.Equal(t, locationURL, location)
	}

	//Error
	{
		mockKeycloakClient.EXPECT().CreateClientRole(accessToken, realmName, clientID, gomock.Any()).Return("", fmt.Errorf("Unexpected error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.CreateClientRole(ctx, "master", clientID, api.RoleRepresentation{})

		assert.NotNil(t, err)
	}
}

func TestGetRealmCustomConfiguration(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmID = "master_id"

	// Get existing config
	{
		var id = realmID
		var keycloakVersion = "4.8.3"
		var realm = "master"
		var displayName = "Master"
		var enabled = true

		var kcRealmRep = kc.RealmRepresentation{
			Id:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)

		var clientID = "ClientID"
		var redirectURI = "http://redirect.url.com/test"

		var realmConfig = dto.RealmConfiguration{
			DefaultClientID:    &clientID,
			DefaultRedirectURI: &redirectURI,
		}

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockConfigurationDBModule.EXPECT().GetConfiguration(ctx, realmID).Return(realmConfig, nil).Times(1)

		configJSON, err := managementComponent.GetRealmCustomConfiguration(ctx, realmID)

		assert.Nil(t, err)
		assert.Equal(t, *configJSON.DefaultClientID, *realmConfig.DefaultClientID)
		assert.Equal(t, *configJSON.DefaultRedirectURI, *realmConfig.DefaultRedirectURI)
	}

	// Get empty config
	{
		var id = realmID
		var keycloakVersion = "4.8.3"
		var realm = "master"
		var displayName = "Master"
		var enabled = true

		var kcRealmRep = kc.RealmRepresentation{
			Id:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockConfigurationDBModule.EXPECT().GetConfiguration(ctx, realmID).Return(dto.RealmConfiguration{}, keycloakb.MissingRealmConfigurationErr{}).Times(1)

		configJSON, err := managementComponent.GetRealmCustomConfiguration(ctx, realmID)

		assert.Nil(t, err)
		assert.Equal(t, *configJSON.DefaultClientID, *new(string))
		assert.Equal(t, *configJSON.DefaultRedirectURI, *new(string))
	}

	// Unknown realm
	{
		var id = realmID
		var keycloakVersion = "4.8.3"
		var realm = "master"
		var displayName = "Master"
		var enabled = true

		var kcRealmRep = kc.RealmRepresentation{
			Id:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, errors.New("error")).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)

		_, err := managementComponent.GetRealmCustomConfiguration(ctx, realmID)

		assert.NotNil(t, err)
	}

	// DB error
	{
		var id = realmID
		var keycloakVersion = "4.8.3"
		var realm = "master"
		var displayName = "Master"
		var enabled = true

		var kcRealmRep = kc.RealmRepresentation{
			Id:              &id,
			KeycloakVersion: &keycloakVersion,
			Realm:           &realm,
			DisplayName:     &displayName,
			Enabled:         &enabled,
		}

		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)

		var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
		mockConfigurationDBModule.EXPECT().GetConfiguration(ctx, realmID).Return(dto.RealmConfiguration{}, errors.New("error")).Times(1)

		_, err := managementComponent.GetRealmCustomConfiguration(ctx, realmID)

		assert.NotNil(t, err)
	}
}

func TestUpdateRealmCustomConfiguration(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockKeycloakClient = mock.NewKeycloakClient(mockCtrl)
	var mockEventDBModule = mock.NewEventDBModule(mockCtrl)
	var mockConfigurationDBModule = mock.NewConfigurationDBModule(mockCtrl)
	var mockLogger = log.NewNopLogger()

	var managementComponent = NewComponent(mockKeycloakClient, mockEventDBModule, mockConfigurationDBModule, mockLogger)

	var accessToken = "TOKEN=="
	var realmID = "master_id"

	var id = realmID
	var keycloakVersion = "4.8.3"
	var realm = "master"
	var displayName = "Master"
	var enabled = true

	var kcRealmRep = kc.RealmRepresentation{
		Id:              &id,
		KeycloakVersion: &keycloakVersion,
		Realm:           &realm,
		DisplayName:     &displayName,
		Enabled:         &enabled,
	}

	var clients = make([]kc.ClientRepresentation, 2)
	var clientID1 = "clientID1"
	var clientName1 = "clientName1"
	var redirectURIs1 = []string{"https://www.cloudtrust.io/*", "https://www.cloudtrust-old.com/*"}
	var clientID2 = "clientID2"
	var clientName2 = "clientName2"
	var redirectURIs2 = []string{"https://www.cloudtrust2.io/*", "https://www.cloudtrust2-old.com/*"}
	clients[0] = kc.ClientRepresentation{
		ClientId:     &clientID1,
		Name:         &clientName1,
		RedirectUris: &redirectURIs1,
	}
	clients[1] = kc.ClientRepresentation{
		ClientId:     &clientID2,
		Name:         &clientName2,
		RedirectUris: &redirectURIs2,
	}

	var ctx = context.WithValue(context.Background(), cs.CtContextAccessToken, accessToken)
	var clientID = "clientID1"
	var redirectURI = "https://www.cloudtrust.io/test"
	var configInit = api.RealmCustomConfiguration{
		DefaultClientID:    &clientID,
		DefaultRedirectURI: &redirectURI,
	}

	// Update config with appropriate values
	{
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetClients(accessToken, realmID).Return(clients, nil).Times(1)
		mockConfigurationDBModule.EXPECT().StoreOrUpdate(ctx, realmID, gomock.Any()).Return(nil).Times(1)
		err := managementComponent.UpdateRealmCustomConfiguration(ctx, realmID, configInit)

		assert.Nil(t, err)
	}

	// Update config with unknown client ID
	{
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetClients(accessToken, realmID).Return(clients, nil).Times(1)

		var clientID = "clientID1Nok"
		var redirectURI = "https://www.cloudtrust.io/test"
		var configInit = api.RealmCustomConfiguration{
			DefaultClientID:    &clientID,
			DefaultRedirectURI: &redirectURI,
		}
		err := managementComponent.UpdateRealmCustomConfiguration(ctx, realmID, configInit)

		assert.NotNil(t, err)
		assert.IsType(t, commonhttp.Error{}, err)
		e := err.(commonhttp.Error)
		assert.Equal(t, 400, e.Status)
	}

	// Update config with invalid redirect URI
	{
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetClients(accessToken, realmID).Return(clients, nil).Times(1)

		var clientID = "clientID1"
		var redirectURI = "https://www.cloudtrustnok.io/test"
		var configInit = api.RealmCustomConfiguration{
			DefaultClientID:    &clientID,
			DefaultRedirectURI: &redirectURI,
		}
		err := managementComponent.UpdateRealmCustomConfiguration(ctx, realmID, configInit)

		assert.NotNil(t, err)
		assert.IsType(t, commonhttp.Error{}, err)
		e := err.(commonhttp.Error)
		assert.Equal(t, 400, e.Status)
	}

	// Update config with invalid redirect URI (trying to take advantage of the dots in the url)
	{
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetClients(accessToken, realmID).Return(clients, nil).Times(1)

		var clientID = "clientID1"
		var redirectURI = "https://wwwacloudtrust.io/test"
		var configInit = api.RealmCustomConfiguration{
			DefaultClientID:    &clientID,
			DefaultRedirectURI: &redirectURI,
		}
		err := managementComponent.UpdateRealmCustomConfiguration(ctx, realmID, configInit)

		assert.NotNil(t, err)
		assert.IsType(t, commonhttp.Error{}, err)
		e := err.(commonhttp.Error)
		assert.Equal(t, 400, e.Status)
	}

	// error while calling GetClients
	{
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kcRealmRep, nil).Times(1)
		mockKeycloakClient.EXPECT().GetClients(accessToken, realmID).Return([]kc.ClientRepresentation{}, errors.New("error")).Times(1)
		err := managementComponent.UpdateRealmCustomConfiguration(ctx, realmID, configInit)

		assert.NotNil(t, err)
	}

	// error while calling GetRealm
	{
		mockKeycloakClient.EXPECT().GetRealm(accessToken, realmID).Return(kc.RealmRepresentation{}, errors.New("error")).Times(1)
		err := managementComponent.UpdateRealmCustomConfiguration(ctx, realmID, configInit)

		assert.NotNil(t, err)
	}
}
