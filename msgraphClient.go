package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/javierjait/azuread/util"
	"github.com/open-networks/go-msgraph"
)

var graphUsers struct {
	users     []map[string]interface{}
	timestamp time.Time
}

// getUsersMap
// Obtiene los usuarios en un map
func getUsersMap() ([]map[string]interface{}, error) {
	now := time.Now().UTC()                                                                       // Obtiene horario actual UTC
	if now.Add(time.Duration(-appConfig.Azure.Cache) * time.Minute).After(graphUsers.timestamp) { // Verifica el última llamado con el tiempo de cache
		client, err := graphClient(appConfig.Azure) // Obtiene el cliente

		if err != nil { // Verifica error
			util.AddToLog("error", fmt.Sprint("Credentials are probably wrong or system time is not synced: ", err))
			return nil, err
		}

		/*msgraph.ListWithSelect("ID, businessPhones, displayName, givenName, mail, mobilePhone, preferredLanguage, surname, userPrincipalName, activePhone, employeeId")*/
		users, err := client.ListUsersMap(msgraph.ListWithSelect(generateSelect())) // Obtiene la lista de usuarios con los campos seleccionados
		if err == nil {                                                             // Si no hay error
			graphUsers.users = users   // Guarda usuarios
			graphUsers.timestamp = now // Guarda horario
		}

		return users, err
	}

	return graphUsers.users, nil
}

// generateSelect
// genera el array de campos a buscar
func generateSelect() string {
	selFields := []string{"userPrincipalName"} // Siempre es necesario

	for _, v := range appConfig.mapFields { // iterate los campos
		if !util.FindInSlice(v, selFields) { // Si no lo encuentra
			selFields = append(selFields, v) // Lo agrega
		}
	}

	return strings.Join(selFields, ",") // devuelve un string separado por coma
}

// grachClient
// crea el cliente para msgraph API
func graphClient(config azureConfigType) (*msgraph.GraphClient, error) {
	return msgraph.NewGraphClient(config.TenantID, config.ApplicationID, config.ClientSecret)
}
