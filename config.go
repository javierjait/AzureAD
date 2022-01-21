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
	BindingUser   string `yaml:"BindingUser"`
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
}

// getConfig
// Lee la configuracion del archivo yaml y la valida
func getConfig() (t configType) {
	t = readConfig()

	validateConfig(&t)

	t.mapFields = make(map[string]string)
	for _, v := range t.MapAux {
		for k, m := range v {
			t.mapFields[k] = m
		}
	}

	return
}

// readConfig
// Lee la configuración del archivo yaml y realiza la serialización
func readConfig() (t configType) {
	data, err := os.ReadFile("config.yaml")
	checkError(err)

	err = yaml.Unmarshal([]byte(data), &t)
	checkError(err)

	return
}

// validateConfig
// Valida los datos de la configuración
func validateConfig(t *configType) {
	validateAzure(t)
	validateLDAP(t)
}

// validateAzure
// Valida los datos de azure
func validateAzure(t *configType) {
	if t.Azure.ApplicationID == "" { // Si ApplicationID es vacío, indica error
		checkError(errors.New("invalid applicationId"))
	}

	if t.Azure.TenantID == "" || t.Azure.ClientSecret == "" { // Si TenantID o ClientSecret es vacío, indica error
		checkError(errors.New("invalid tenant or client secret"))
	}

	if t.Azure.Cache < 0 { // El cache tiene que ser mayor o igual que 0
		t.Azure.Cache = 0
	}

	if t.Azure.Cache > 1440 { // El cache tiene que ser menor a 1440 minutos
		t.Azure.Cache = 1440
	}

	_, err := graphClient(t.Azure) // Valida el cliente azure
	checkError(err)
}

// validateLDAP
// validad los datos del servidor LDAP
func validateLDAP(t *configType) {
	// Check LDAP port
	if t.LDAP.Port == 0 { // Si el puerto es 0, pone los puertos predeterminados
		if t.LDAP.Secure {
			t.LDAP.Port = 636 // Puerto seguro predeterminado
		} else {
			t.LDAP.Port = 389 // Puerto básico predeterminado
		}
	}

	if t.LDAP.Secure {
		// Check directory
		if !strings.HasSuffix(t.LDAP.CertificateDirectory, "/") && !strings.HasSuffix(t.LDAP.CertificateDirectory, "\\") {
			t.LDAP.CertificateDirectory += "/"
		}

		if _, err := os.Stat(t.LDAP.CertificateDirectory); err != nil { // Verifica que la carpeta exista
			checkError(errors.New("certificate directory not found"))
		}

		if _, err := os.Stat(t.LDAP.CertificateDirectory + "cert.pem"); err != nil { // Verifica que el archivo cert.pem exista
			checkError(errors.New("certificate not found"))
		}

		if _, err := os.Stat(t.LDAP.CertificateDirectory + "key.pem"); err != nil { // Verifica que el archivo key.pem exista
			checkError(errors.New("private key not found"))
		}
	}
}

// checkError
// Agrega el error al log, lo muestra y cierra la aplicación
func checkError(err error) {
	if err != nil { // si hay error
		util.AddToLog("error", err.Error()) // Lo agrega al archivo de log
		log.Fatal(err)                      // Muestra el error en consola y cierra
	}
}
