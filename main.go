package main

import (
	"fmt"
	"log"

	"github.com/nmcclain/ldap"
)

var appConfig configType

func main() {
	appConfig = getConfig()

	s := ldap.NewServer()
	handler := ldapHandler{}
	s.BindFunc("", handler)
	s.SearchFunc("", handler)
	s.UnbindFunc("", handler)
	s.CloseFunc("", handler)

	if appConfig.LDAP.Secure {
		if err := s.ListenAndServeTLS(fmt.Sprintf(":%d", appConfig.LDAP.Port), "cert.pem", "key.pem"); err != nil {
			log.Fatalf("LDAP Server Failed: %s", err.Error())
		}
	} else {
		if err := s.ListenAndServe(fmt.Sprintf(":%d", appConfig.LDAP.Port)); err != nil {
			log.Fatalf("LDAP Server Failed: %s", err.Error())
		}
	}
}
