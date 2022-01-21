package main

import (
	"context"
	"errors"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

var cacheAccessor = &TokenCache{}

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
	//log.Printf("Start: %s", time.Now())
	app, err := azureClient(appConfig.Azure.ApplicationID, appConfig.Azure.BindingUser == user)
	if err != nil {
		return "", err
	}

	var result public.AuthResult

	err = errors.New("access token not found")

	if appConfig.Azure.BindingUser == user {
		// look in the cache to see if the account to use has been cached
		accounts := app.Accounts()
		if len(accounts) > 0 {
			var userAccount public.Account

			for _, account := range accounts {
				if account.PreferredUsername == user {
					userAccount = account
				}
			}

			result, err = app.AcquireTokenSilent(
				context.Background(),
				[]string{"user.read"},
				public.WithSilentAccount(userAccount),
			)
		}
	}

	if err != nil {
		result, err = app.AcquireTokenByUsernamePassword(
			context.Background(),
			[]string{"user.read"},
			user,
			password,
		)
	}

	//log.Printf("End bind login: %s", time.Now())

	return result.AccessToken, err
}

func azureClient(clientID string, cache bool) (app public.Client, err error) {
	if cache {
		return public.New(clientID, public.WithCache(cacheAccessor), public.WithAuthority("https://login.microsoftonline.com/organizations"))
	}
	return public.New(clientID, public.WithAuthority("https://login.microsoftonline.com/organizations"))
}
