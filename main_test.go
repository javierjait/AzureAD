package main

import "testing"

func TestGetUserName(t *testing.T) {
	expectedUser := "javierjait"
	expectedDomain := "gmail.com"
	user, domain := getUsername("cn=javierjait,dc=gmail,dc=com")
	if user != expectedUser || domain != expectedDomain {
		t.Errorf("Getusername was incorrect, got: %s@%s, want: %s@%s.", user, domain, expectedUser, expectedDomain)
	}
}
