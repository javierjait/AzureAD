package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/javierjait/azuread/util"
	"github.com/nmcclain/ldap"
)

type ldapHandler struct{}

/// Handlers para el servidor LDAP

// Bind handler
func (h ldapHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	username, domain := getUsername(bindDN)
	_, err := azureLogin(username+"@"+domain, bindSimplePw)
	if err == nil {
		util.AddToLog("bind", fmt.Sprintf("dn: %s, mail: %s", bindDN, username+"@"+domain))
		return ldap.LDAPResultSuccess, nil
	}

	util.AddToLog("error", fmt.Sprintf("invalid username or password for: dn: %s, mail: %s", bindDN, username+"@"+domain))
	util.AddToLog("error", err.Error())

	return ldap.LDAPResultInvalidCredentials, nil
}

func (h ldapHandler) Search(boundDN string, searchReq ldap.SearchRequest, conn net.Conn) (ldap.ServerSearchResult, error) {
	if boundDN != "" {
		util.AddToLog("search", fmt.Sprintf("%+v", searchReq))

		searchType := searchReq.Filter[1]           // get filter type
		if searchType != '|' && searchType != '&' { // check filter type | &
			searchType = '|'
		}

		filter := buildReqAttributesList(searchReq.Filter, []string{}) // forma el filtro

		usersFound := findUsers(filter, searchType) // busca usuarios

		if len(usersFound) > 0 {
			var entries []*ldap.Entry

			for _, u := range usersFound {
				var userName, domain string

				azureValue, ok := getAzureValue(u, "userPrincipalName") // obtiene el valor azure
				if ok {
					switch upn := azureValue.(type) {
					case string:
						userName, domain = getUsernameDomain(upn) // obtiene el domain del upn
					}
				}

				if userName != "" && domain != "" {
					baseDN := getDN(domain) // genera el DN
					var entry ldap.Entry

					entry.DN = fmt.Sprintf("cn=%s,%s", userName, baseDN)
					for azureField := range u {
						azureValue, ok := getAzureValue(u, azureField)
						if !ok {
							continue
						}

						ldapFields := getLDAPFields(azureField) // Obtiene los datos ldap el campo azure
						for _, ldapName := range ldapFields {
							switch v := azureValue.(type) {
							case string:
								entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapName, Values: []string{v}})
							default:
								if v != nil {
									entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapName, Values: []string{fmt.Sprint(v)}})
								}
							}
						}
					}

					entries = append(entries, &entry)
				}
			}

			//log.Printf("End: %s", time.Now())
			return ldap.ServerSearchResult{Entries: entries, Referrals: []string{}, Controls: []ldap.Control{}, ResultCode: ldap.LDAPResultSuccess}, nil
		} else {
			util.AddToLog("search", "not found")
		}
	} else {
		util.AddToLog("search", "not bound")
	}

	return ldap.ServerSearchResult{Entries: nil, Referrals: []string{}, Controls: []ldap.Control{}, ResultCode: ldap.LDAPResultOperationsError}, nil
}

// getAzureField
// devuelve el valor del campo azure del campo ldap seleccionado
func getAzureField(ldapField string) string {
	for k, v := range appConfig.mapFields { // iterate mapFields
		if strings.EqualFold(ldapField, k) { // verifica si son iguales insensitive
			return v
		}
	}

	return ""
}

// getAzureValue
// devuelve el valor del campo y ok = true si lo encuentra. Si ok es falso, no lo encontró
func getAzureValue(user map[string]interface{}, azureField string) (value interface{}, ok bool) {
	value, ok = user[azureField]

	return
}

// findUsers
// Busca los usuarios según el filtro y el tipo de búsqueda (or/and)
func findUsers(filter []string, searchType byte) (usersFound []map[string]interface{}) {
	users, err := getUsersMap() // Obtiene todos los usuarios
	if err == nil {
		for _, v := range users { // Iterate users
			found := false
			for _, f := range filter { // Iterate filter
				ldapField := ""
				ldapFilter := ""
				ldapAttr := strings.Split(f, "=") // Separa el filtro

				if len(ldapAttr) != 2 { // Si longitud es distinto 2, continua sin buscar
					continue
				}
				ldapField = ldapAttr[0]
				ldapFilter = strings.ToLower(ldapAttr[1])

				azureField := getAzureField(ldapField) // Obtiene el campo azure
				if azureField == "" {
					found = false
					continue
				}
				fieldValueAux, ok := v[azureField] // Obtiene el valor del campo
				if !ok {
					found = false
					continue
				}
				if fieldValueAux == nil { // Si no hay valor found = false y continua
					found = false
					continue
				}

				fieldValue := strings.ToLower(fmt.Sprint(fieldValueAux)) // Lo convierte en minuscula
				lenFilter := len(ldapFilter)
				if lenFilter == 1 && ldapFilter == "*" { // busca todos
					found = fieldValue != ""
				} else if lenFilter > 2 && strings.HasPrefix(ldapFilter, "*") && strings.HasSuffix(ldapFilter, "*") { // Busca prefijos y sufijos
					if lenFilter > 2 {
						// Elimina los asteriscos
						ldapFilter = ldapFilter[:len(ldapFilter)-1]
						ldapFilter = ldapFilter[1:]
						found = strings.Contains(fieldValue, ldapFilter) // Busca si contiene el filtro
					}
				} else if strings.HasPrefix(ldapFilter, "*") { // Busca * en prefijo
					if lenFilter > 1 {
						ldapFilter = ldapFilter[1:]                       // Elimina asterisco
						found = strings.HasSuffix(fieldValue, ldapFilter) // Busca si contiene el filtro
					}
				} else if strings.HasSuffix(ldapFilter, "*") { // Busca * en sufijo
					if lenFilter > 1 {
						ldapFilter = ldapFilter[:len(ldapFilter)-1]       // Elimina asterisco
						found = strings.HasPrefix(fieldValue, ldapFilter) // Busca si contiene el filtro
					}
				} else {
					found = strings.EqualFold(ldapFilter, fieldValue) // Búsqueda igual insensitve
				}

				if searchType == '&' { // si es and
					if !found { // Si found es falso
						break // Finaliza
					}
				} else { // Es or
					if found { // Si found es verdadero
						break // Finaliza
					}
				}
			}
			if found { // Si found es verdadero
				usersFound = append(usersFound, v) // Agrega al array de usuarios encontrados
			}
		}
	}

	return
}

// getLDAPFields
// Devuelve el array de los campos ldap según el campo azure
func getLDAPFields(azureField string) (ldapFields []string) {
	for k, v := range appConfig.mapFields {
		if v == azureField {
			ldapFields = append(ldapFields, k)
		}
	}

	return
}

// Unbind handler
func (h ldapHandler) Unbind(boundDN string, conn net.Conn) (ldap.LDAPResultCode, error) {
	util.AddToLog("close", fmt.Sprintf("Unbind: %s", boundDN))

	return ldap.LDAPResultSuccess, nil
}

// Close handler
func (h ldapHandler) Close(boundDN string, conn net.Conn) error {
	if boundDN != "" {
		util.AddToLog("close", boundDN)
	}

	return nil
}

func buildReqAttributesList(filter string, filters []string) []string {
	maxp := len(filter)
	start := -1
	descended := false
	for p, c := range filter {
		if c == '(' {
			if p+1 < maxp {
				start = p + 1
			}
		} else if c == ')' {
			if start > -1 {
				descended = true
				filters = buildReqAttributesList(filter[start:p], filters)
			}
			start = -1
		}
	}

	if !descended {
		filters = append(filters, filter)
	}
	return filters
}

// getDN
func getDN(domain string) (baseDN string) {
	if domain != "" { // Si el domain no está vacío
		domainParts := strings.Split(domain, ".") // Genera un array de strings separados por "."
		for _, v := range domainParts {           // Iterate
			baseDN += fmt.Sprintf("dc=%s,", v) // Agrega dc=v,
		}
		baseDN = strings.TrimSuffix(baseDN, ",") // Remueve la última coma
	}

	return
}

// getUsernameDomain
// Devuelve el domain del mail
func getUsernameDomain(mail string) (user, domain string) {
	mailSlice := strings.Split(mail, "@")                                // Genera un array de strings separados por "@"
	if len(mailSlice) == 2 && mailSlice[0] != "" && mailSlice[1] != "" { // Verifica longitud y que contengan valores
		user = mailSlice[0]
		domain = mailSlice[1]
	}

	return
}
