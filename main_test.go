package main

import "testing"

func TestGetUserName(t *testing.T) {
	/*expectedUser := "javierjait"
	expectedDomain := "gmail.com"
	user, domain := getUsername("cn=javierjait,dc=gmail,dc=com")*/
	expectedUser := "terencio"
	expectedDomain := "LesRoches.ES"
	user, domain := getUsername("cn=terencio,O=LesRoches,C=ES")
	if user != expectedUser || domain != expectedDomain {
		t.Errorf("Getusername was incorrect, got: %s@%s, want: %s@%s.", user, domain, expectedUser, expectedDomain)
	}
}

func TestGetDN(t *testing.T) {
	expectedDN := "dc=gmail,dc=com"
	baseDN := getDN("gmail.com")
	//expectedDN := ""
	//baseDN := getDN("")
	if baseDN != expectedDN {
		t.Errorf("getDN was incorrect. Expected: %s, Got: %s", expectedDN, baseDN)
	}
}

func TestGetConfig(t *testing.T) {
	appConfig = getConfig()
	if appConfig.Azure.ApplicationID == "" {
		t.Errorf("GetConfig failed. AppConfig %+v", appConfig)
	}
}
