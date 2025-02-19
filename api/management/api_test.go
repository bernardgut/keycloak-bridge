package management_api

import (
	"encoding/json"
	"strings"
	"testing"

	kc "github.com/cloudtrust/keycloak-client"
	"github.com/stretchr/testify/assert"
)

func TestConvertCredential(t *testing.T) {
	var credKc kc.CredentialRepresentation
	var credType = "password"
	var credID = "123456"
	var configKc = "{}"

	credKc.Type = &credType
	credKc.Id = &credID
	credKc.CredentialData = nil

	assert.Equal(t, credKc.Type, ConvertCredential(&credKc).Type)
	assert.Equal(t, credKc.Id, ConvertCredential(&credKc).ID)
	assert.Nil(t, ConvertCredential(&credKc).CredentialData)

	credKc.CredentialData = &configKc
	assert.NotNil(t, ConvertCredential(&credKc).CredentialData)
	assert.Equal(t, "{}", *ConvertCredential(&credKc).CredentialData)
}

func TestConvertToAPIUser(t *testing.T) {
	var kcUser kc.UserRepresentation
	m := make(map[string][]string)

	// Phone number
	assert.Nil(t, ConvertToAPIUser(kcUser).PhoneNumber)
	kcUser.Attributes = &m
	m["phoneNumber"] = []string{"+4122555555"}
	assert.NotNil(t, ConvertToAPIUser(kcUser).PhoneNumber)

	// Label
	assert.Nil(t, ConvertToAPIUser(kcUser).Label)
	kcUser.Attributes = &m
	m["label"] = []string{"a label"}
	assert.NotNil(t, ConvertToAPIUser(kcUser).Label)

	// Gender
	assert.Nil(t, ConvertToAPIUser(kcUser).Gender)
	kcUser.Attributes = &m
	m["gender"] = []string{"a gender"}
	assert.NotNil(t, ConvertToAPIUser(kcUser).Gender)

	// Birthdate
	assert.Nil(t, ConvertToAPIUser(kcUser).BirthDate)
	kcUser.Attributes = &m
	m["birthDate"] = []string{"25/12/0"}
	assert.NotNil(t, ConvertToAPIUser(kcUser).BirthDate)

	// PhoneNumberVerified
	assert.Nil(t, ConvertToAPIUser(kcUser).PhoneNumberVerified)
	kcUser.Attributes = &m
	m["phoneNumberVerified"] = []string{"true"}
	assert.True(t, *ConvertToAPIUser(kcUser).PhoneNumberVerified)

	// Locale
	assert.Nil(t, ConvertToAPIUser(kcUser).Locale)
	kcUser.Attributes = &m
	m["locale"] = []string{"en"}
	assert.NotNil(t, *ConvertToAPIUser(kcUser).Locale)
}

func TestConvertToAPIUsersPage(t *testing.T) {
	var count = 10
	var input = kc.UsersPageRepresentation{Count: &count, Users: []kc.UserRepresentation{kc.UserRepresentation{}, kc.UserRepresentation{}}}
	var output = ConvertToAPIUsersPage(input)
	assert.Equal(t, count, *output.Count)
	assert.Equal(t, len(input.Users), len(output.Users))
}

func TestConvertToAPIUsersPageEmptySet(t *testing.T) {
	var input = kc.UsersPageRepresentation{Count: nil, Users: nil}
	var output = ConvertToAPIUsersPage(input)
	assert.NotNil(t, output.Users)
	assert.NotNil(t, output.Count)
	assert.Equal(t, 0, len(output.Users))
	assert.Equal(t, 0, *output.Count)

	var jsonRaw, _ = json.Marshal(output)
	var json = string(jsonRaw)
	assert.True(t, strings.Contains(json, `"users":[]`))
	assert.True(t, strings.Contains(json, `"count":0`))
}

func TestConvertToKCUser(t *testing.T) {
	var user UserRepresentation

	// Phone number
	assert.Nil(t, ConvertToKCUser(user).Attributes)
	var phoneNumber = "+4122555555"
	user.PhoneNumber = &phoneNumber
	assert.Equal(t, phoneNumber, (*ConvertToKCUser(user).Attributes)["phoneNumber"][0])

	// Label
	var label = "a label"
	user.Label = &label
	assert.Equal(t, label, (*ConvertToKCUser(user).Attributes)["label"][0])

	// Gender
	var gender = "a gender"
	user.Gender = &gender
	assert.Equal(t, gender, (*ConvertToKCUser(user).Attributes)["gender"][0])

	// Birthdate
	var date = "25/12/0"
	user.BirthDate = &date
	assert.Equal(t, date, (*ConvertToKCUser(user).Attributes)["birthDate"][0])

	// PhoneNumberVerified
	var verified = true
	user.PhoneNumberVerified = &verified
	assert.Equal(t, "true", (*ConvertToKCUser(user).Attributes)["phoneNumberVerified"][0])

	// Locale
	var locale = "it"
	user.Locale = &locale
	assert.Equal(t, locale, (*ConvertToKCUser(user).Attributes)["locale"][0])
}

func TestValidateUserRepresentation(t *testing.T) {
	{
		user := createValidUserRepresentation()
		assert.Nil(t, user.Validate())
	}

	id := "#12345"
	username := "username!"
	email := "usernamcompany.com"
	phoneNumber := "415174234"
	label := ""
	gender := "Male"
	birthDate := "1990-13-28"
	locale := "english"
	groups := []string{"f467ed7c", "7767ed7c-0a1d-4eee-9bb8-669c6f89c007"}
	roles := []string{"abcded7", "7767ed7c-0a1d-4eee-9bb8-669c6f898888"}
	empty := ""

	var users []UserRepresentation
	for i := 0; i < 12; i++ {
		users = append(users, createValidUserRepresentation())
	}

	users[0].ID = &id
	users[1].Username = &username
	users[2].Email = &email
	users[3].PhoneNumber = &phoneNumber
	users[4].Label = &label
	users[5].Gender = &gender
	users[6].BirthDate = &birthDate
	users[7].Groups = &groups
	users[8].Roles = &roles
	users[9].Locale = &locale
	users[10].FirstName = &empty
	users[11].LastName = &empty

	for _, user := range users {
		assert.NotNil(t, user.Validate())
	}
}

func TestValidateRoleRepresentation(t *testing.T) {
	{
		role := createValidRoleRepresentation()
		assert.Nil(t, role.Validate())
	}

	id := "f467ed7c"
	name := "name *"
	description := ""

	var roles []RoleRepresentation
	for i := 0; i < 4; i++ {
		roles = append(roles, createValidRoleRepresentation())
	}

	roles[0].ID = &id
	roles[1].Name = &name
	roles[2].Description = &description
	roles[3].ContainerID = &id

	for _, role := range roles {
		assert.NotNil(t, role.Validate())
	}
}

func TestValidatePasswordRepresentation(t *testing.T) {
	{
		password := createValidPasswordRepresentation()
		assert.Nil(t, password.Validate())
	}

	value := ""
	password := createValidPasswordRepresentation()
	password.Value = &value

	assert.NotNil(t, password.Validate())

}

func TestValidateRealmCustomConfiguration(t *testing.T) {
	{
		config := createValidRealmCustomConfiguration()
		assert.Nil(t, config.Validate())
	}

	defaultClientID := "something$invalid"
	defaultRedirectURI := "ht//tp://company.com"

	var configs []RealmCustomConfiguration
	for i := 0; i < 2; i++ {
		configs = append(configs, createValidRealmCustomConfiguration())
	}

	configs[0].DefaultClientID = &defaultClientID
	configs[1].DefaultRedirectURI = &defaultRedirectURI

	for _, config := range configs {
		assert.NotNil(t, config.Validate())
	}
}

func TestValidateRequiredAction(t *testing.T) {
	{
		action := createValidRequiredAction()
		assert.Nil(t, action.Validate())
	}

	action := RequiredAction("^")
	assert.NotNil(t, action.Validate())
}

func createValidUserRepresentation() UserRepresentation {
	var groups = []string{"f467ed7c-0a1d-4eee-9bb8-669c6f89c0ee", "7767ed7c-0a1d-4eee-9bb8-669c6f89c007"}
	var roles = []string{"abcded7c-0a1d-4eee-9bb8-669c6f89c0ee", "7767ed7c-0a1d-4eee-9bb8-669c6f898888"}

	id := "f467ed7c-0a1d-4eee-9bb8-669c6f89c0ee"
	username := "username"
	email := "username@company.com"
	boolTrue := true
	phoneNumber := "+415174234"
	firstName := "Firstname"
	lastName := "Lastname"
	label := "label"
	gender := "F"
	birthDate := "1990-12-28"
	locale := "en"

	var user = UserRepresentation{}
	user.ID = &id
	user.Username = &username
	user.Email = &email
	user.Enabled = &boolTrue
	user.EmailVerified = &boolTrue
	user.PhoneNumber = &phoneNumber
	user.PhoneNumberVerified = &boolTrue
	user.FirstName = &firstName
	user.LastName = &lastName
	user.Label = &label
	user.Gender = &gender
	user.BirthDate = &birthDate
	user.Groups = &groups
	user.Roles = &roles
	user.Locale = &locale

	return user
}

func createValidRoleRepresentation() RoleRepresentation {
	id := "f467ed7c-0a1d-4eee-9bb8-669c6f89c0ee"
	name := "name"
	description := "description"
	boolTrue := true

	var role = RoleRepresentation{}
	role.ID = &id
	role.Name = &name
	role.Description = &description
	role.ContainerID = &id
	role.ClientRole = &boolTrue
	role.Composite = &boolTrue

	return role
}

func createValidPasswordRepresentation() PasswordRepresentation {
	password := "password"

	return PasswordRepresentation{
		Value: &password,
	}
}

func createValidRealmCustomConfiguration() RealmCustomConfiguration {
	defaultClientID := "backofficeid"
	defaultRedirectURI := "http://company.com"

	return RealmCustomConfiguration{
		DefaultClientID:    &defaultClientID,
		DefaultRedirectURI: &defaultRedirectURI,
	}
}

func createValidRequiredAction() RequiredAction {
	return RequiredAction("verify-email")
}
