package main

import (
	"testing"
)

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

func TestGetUsernameDomain(t *testing.T) {
	type gud struct {
		mail   string
		user   string
		domain string
	}
	var testGud []gud
	testGud = append(testGud, gud{"", "", ""})
	testGud = append(testGud, gud{"javierjait@yahoo.com", "javierjait", "yahoo.com"})
	testGud = append(testGud, gud{"test@test.com", "test", "test.com"})
	testGud = append(testGud, gud{"@test.com", "", ""})
	testGud = append(testGud, gud{"test.com", "", ""})

	for _, v := range testGud {
		u, d := getUsernameDomain(v.mail)
		if u != v.user || d != v.domain {
			t.Errorf("getUsernameDomain error. Expected u: %s, d: %s. Got u: %s, d: %s", v.user, v.domain, u, d)
		}
	}
}
