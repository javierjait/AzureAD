package main

import (
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/javierjait/azuread/util"
)

type TokenCache struct {
	cache []byte // Estructura del cache en memoria
}

// Reemplaza los datos del cache
func (t *TokenCache) Replace(cache cache.Unmarshaler, key string) {
	err := cache.Unmarshal(t.cache)
	if err != nil {
		util.AddToLog("error", err.Error())
	}
}

// Exporta los datos del cache
func (t *TokenCache) Export(cache cache.Marshaler, key string) {
	data, err := cache.Marshal()
	if err != nil {
		util.AddToLog("error", err.Error())
	}
	t.cache = data
}
