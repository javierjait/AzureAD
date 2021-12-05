package main

import (
	"context"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

func getUsername(bindDN string) (username, domain string) {
	bindArray := strings.Split(bindDN, ",")
	for _, v := range bindArray {
		if strings.HasPrefix(strings.ToLower(v), "cn=") {
			username = strings.Replace(v, "cn=", "", 1)
		}
		if strings.HasPrefix(strings.ToLower(v), "dc=") || strings.HasPrefix(strings.ToLower(v), "o=") || strings.HasPrefix(strings.ToLower(v), "c=") {
			equalSign := strings.Index(v, "=")
			if equalSign != -1 {
				domain += v[equalSign+1:] + "."
			}
		}
	}

	domain = strings.TrimRight(domain, ".")

	return
}

func azureLogin(user, password string) (string, error) {
	app, err := azureClient(appConfig.Azure.ApplicationID)
	if err != nil {
		return "", err
	}

	result, err := app.AcquireTokenByUsernamePassword(
		context.Background(),
		[]string{"user.read"},
		user,
		password,
	)

	return result.AccessToken, err
}

func azureClient(clientID string) (app public.Client, err error) {
	return public.New(clientID, public.WithAuthority("https://login.microsoftonline.com/organizations"))
}
