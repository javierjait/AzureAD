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

func getUsersMap() ([]map[string]interface{}, error) {
	now := time.Now().UTC()
	if now.Add(time.Duration(-appConfig.Azure.Cache) * time.Minute).After(graphUsers.timestamp) {
		client, err := graphClient(appConfig.Azure)

		if err != nil {
			util.AddToLog("error", fmt.Sprint("Credentials are probably wrong or system time is not synced: ", err))
			return nil, err
		}

		/*msgraph.ListWithSelect("ID, businessPhones, displayName, givenName, mail, mobilePhone, preferredLanguage, surname, userPrincipalName, activePhone, employeeId")*/
		users, err := client.ListUsersMap(msgraph.ListWithSelect(generateSelect()))
		if err == nil {
			graphUsers.users = users
			graphUsers.timestamp = now
		}

		return users, err
	}

	return graphUsers.users, nil
}

func generateSelect() string {
	selFields := []string{"userPrincipalName"}

	for _, v := range appConfig.mapFields {
		if !util.FindInSlice(v, selFields) {
			selFields = append(selFields, v)

		}
	}

	return strings.Join(selFields, ",")
}

func graphClient(config azureConfigType) (*msgraph.GraphClient, error) {
	return msgraph.NewGraphClient(config.TenantID, config.ApplicationID, config.ClientSecret)
}
