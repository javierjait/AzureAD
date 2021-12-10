package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/javierjait/azuread/util"
	"gopkg.in/yaml.v2"
)

type azureConfigType struct {
	TenantID      string `yaml:"TenantId"`
	ApplicationID string `yaml:"ApplicationId"`
	ClientSecret  string `yaml:"ClientSecret"`
	Cache         int    `yaml:"CacheMinutes"`
}

type ldapConfigType struct {
	Secure               bool   `yaml:"Secure"`
	Port                 int    `yaml:"Port"`
	CertificateDirectory string `yaml:"CertificateDirectory"`
	Domain               string `yaml:"Domain"`
}

type configType struct {
	Azure     azureConfigType     `yaml:"Azure"`
	LDAP      ldapConfigType      `yaml:"LDAP"`
	MapAux    []map[string]string `yaml:"Mapping"`
	mapFields map[string]string   // key ldapfield, value azureField
	//mapFields map[string][]string
}

func getConfig() (t configType) {
	t = readConfig()

	validateConfig(&t)

	t.mapFields = make(map[string]string)
	//t.mapFields = make(map[string][]string)
	for _, v := range t.MapAux {
		for k, m := range v {
			t.mapFields[k] = m
			//t.mapFields[k] = append(t.mapFields[k], m)
		}
	}

	return
}

func readConfig() (t configType) {
	data, err := os.ReadFile("config.yaml")
	checkError(err)

	err = yaml.Unmarshal([]byte(data), &t)
	checkError(err)

	return
}

func validateConfig(t *configType) {
	validateAzure(t)
	validateLDAP(t)
}

func validateAzure(t *configType) {
	if t.Azure.ApplicationID == "" {
		checkError(errors.New("invalid applicationId"))
	}

	if t.Azure.TenantID == "" || t.Azure.ClientSecret == "" {
		checkError(errors.New("invalid tenant or client secret"))
	}

	if t.Azure.Cache < 0 {
		t.Azure.Cache = 0
	}

	if t.Azure.Cache > 1440 {
		t.Azure.Cache = 1440
	}

	_, err := graphClient(t.Azure)
	checkError(err)
}

func validateLDAP(t *configType) {
	// Check LDAP port
	if t.LDAP.Port == 0 {
		if t.LDAP.Secure {
			t.LDAP.Port = 636
		} else {
			t.LDAP.Port = 389
		}
	}

	if t.LDAP.Secure {
		// Check directory
		if !strings.HasSuffix(t.LDAP.CertificateDirectory, "/") && !strings.HasSuffix(t.LDAP.CertificateDirectory, "\\") {
			t.LDAP.CertificateDirectory += "/"
		}

		if _, err := os.Stat(t.LDAP.CertificateDirectory); err != nil {
			checkError(errors.New("certificate directory not found"))
		}

		if _, err := os.Stat(t.LDAP.CertificateDirectory + "cert.pem"); err != nil {
			checkError(errors.New("certificate not found"))
		}

		if _, err := os.Stat(t.LDAP.CertificateDirectory + "key.pem"); err != nil {
			checkError(errors.New("private key not found"))
		}
	}
}

func checkError(err error) {
	if err != nil {
		util.AddToLog("error", err.Error())
		log.Fatal(err)
	}
}
