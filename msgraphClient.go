package main

import (
	"fmt"
	"time"

	"github.com/javierjait/azuread/util"
	"github.com/open-networks/go-msgraph"
)

var graphUsers struct {
	users     []map[string]interface{}
	timestamp time.Time
}

/*func getUsers() (msgraph.Users, error) {
	client, err := graphClient(appConfig.Azure)

	if err != nil {
		util.AddToLog("error", fmt.Sprint("Credentials are probably wrong or system time is not synced: ", err))
		return msgraph.Users{}, err
	}

	return client.ListUsers(msgraph.ListWithSelect("ID, businessPhones, displayName, givenName, mail, mobilePhone, preferredLanguage, surname, userPrincipalName, activePhone, employeeId"))
}

func getUser(id string) (msgraph.User, error) {
	users, err := getUsers()
	if err == nil {
		for _, v := range users {
			if id == v.EmployeeID {
				return v, nil
			}
		}
	}

	return msgraph.User{}, errors.New("not found")
}*/

func getUsersMap() ([]map[string]interface{}, error) {
	now := time.Now().UTC()
	//log.Printf("Now: %s, Check: %s, Timestamp: %s", now, now.Add(time.Duration(-appConfig.Azure.Cache)*time.Minute), graphUsers.timestamp)
	//log.Print(now.Add(time.Duration(-appConfig.Azure.Cache) * time.Minute).After(graphUsers.timestamp))
	if now.Add(time.Duration(-appConfig.Azure.Cache) * time.Minute).After(graphUsers.timestamp) {
		//log.Print("get from API")
		client, err := graphClient(appConfig.Azure)

		if err != nil {
			util.AddToLog("error", fmt.Sprint("Credentials are probably wrong or system time is not synced: ", err))
			return nil, err
		}

		users, err := client.ListUsersMap(msgraph.ListWithSelect("ID, businessPhones, displayName, givenName, mail, mobilePhone, preferredLanguage, surname, userPrincipalName, activePhone, employeeId"))
		if err == nil {
			graphUsers.users = users
			graphUsers.timestamp = now
		}

		return users, err
	}

	//log.Print("get from memory")

	return graphUsers.users, nil
}

func graphClient(config azureConfigType) (*msgraph.GraphClient, error) {
	return msgraph.NewGraphClient(config.TenantID, config.ApplicationID, config.ClientSecret)
}
