package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/javierjait/azuread/util"
	"github.com/nmcclain/ldap"
)

type ldapHandler struct{}

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

		filter := buildReqAttributesList(searchReq.Filter, []string{})

		usersFound := findUsers(filter, searchType)

		if len(usersFound) > 0 {
			var entries []*ldap.Entry

			for _, u := range usersFound {
				var userName, domain string

				azureValue, ok := getAzureValue(u, "userPrincipalName")
				if ok {
					switch upn := azureValue.(type) {
					case string:
						userName, domain = getUsernameDomain(upn)
					}
				}

				if userName != "" && domain != "" {
					baseDN := getDN(domain)
					var entry ldap.Entry

					entry.DN = fmt.Sprintf("cn=%s,%s", userName, baseDN)
					for azureField := range u {
						ldapFields := getLDAPFields(azureField)
						azureValue, _ := getAzureValue(u, azureField)
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

			return ldap.ServerSearchResult{Entries: entries, Referrals: []string{}, Controls: []ldap.Control{}, ResultCode: ldap.LDAPResultSuccess}, nil
		} else {
			util.AddToLog("search", "not found")
		}
	} else {
		util.AddToLog("search", "not bound")
	}

	return ldap.ServerSearchResult{Entries: nil, Referrals: []string{}, Controls: []ldap.Control{}, ResultCode: ldap.LDAPResultOperationsError}, nil
}

func getAzureField(ldapField string) string {
	for k, v := range appConfig.mapFields {
		if strings.EqualFold(ldapField, k) {
			return v
		}
	}

	return ""
}

func getAzureValue(user map[string]interface{}, azureField string) (value interface{}, ok bool) {
	value, ok = user[azureField]

	return
}

func findUsers(filter []string, searchType byte) (usersFound []map[string]interface{}) {
	users, err := getUsersMap()
	if err == nil {
		for _, v := range users {
			found := false
			for _, f := range filter {
				ldapField := ""
				ldapFilter := ""
				ldapAttr := strings.Split(f, "=")

				if len(ldapAttr) != 2 {
					continue
				}
				ldapField = ldapAttr[0]
				ldapFilter = strings.ToLower(ldapAttr[1])

				azureField := getAzureField(ldapField)
				if azureField == "" {
					if searchType == '&' {
						found = false
						break
					} else {
						continue
					}
				}

				fieldValueAux, ok := v[azureField]
				if !ok {
					continue
				}

				if fieldValueAux == nil {
					continue
				}
				fieldValue := strings.ToLower(fmt.Sprint(fieldValueAux))

				lenFilter := len(ldapFilter)
				if lenFilter == 1 && ldapFilter == "*" {
					found = fieldValue != ""
				} else if lenFilter > 2 && strings.HasPrefix(ldapFilter, "*") && strings.HasSuffix(ldapFilter, "*") {
					if lenFilter > 2 {
						ldapFilter = ldapFilter[:len(ldapFilter)-1]
						ldapFilter = ldapFilter[1:]
						found = strings.Contains(fieldValue, ldapFilter)
					}
				} else if strings.HasPrefix(ldapFilter, "*") {
					if lenFilter > 1 {
						ldapFilter = ldapFilter[1:]
						found = strings.HasSuffix(fieldValue, ldapFilter)
					}
				} else if strings.HasSuffix(ldapFilter, "*") {
					if lenFilter > 1 {
						ldapFilter = ldapFilter[:len(ldapFilter)-1]
						found = strings.HasPrefix(fieldValue, ldapFilter)
					}
				} else {
					found = strings.EqualFold(ldapFilter, fieldValue)
				}

				if searchType == '&' {
					if !found {
						break
					}
				} else {
					if found {
						break
					}
				}
			}
			if found {
				usersFound = append(usersFound, v)
			}
		}
	}

	return
}

func getLDAPFields(azureField string) (ldapFields []string) {
	for k, v := range appConfig.mapFields {
		if v == azureField {
			ldapFields = append(ldapFields, k)
		}
	}

	return
}

func (h ldapHandler) Unbind(boundDN string, conn net.Conn) (ldap.LDAPResultCode, error) {
	util.AddToLog("close", fmt.Sprintf("Unbind: %s", boundDN))

	return ldap.LDAPResultSuccess, nil
}

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

func getDN(domain string) (baseDN string) {
	if domain != "" {
		domainParts := strings.Split(domain, ".")
		for _, v := range domainParts {
			baseDN += fmt.Sprintf("dc=%s,", v)
		}
		baseDN = strings.TrimSuffix(baseDN, ",")
	}

	return
}

func getUsernameDomain(mail string) (user, domain string) {
	mailSlice := strings.Split(mail, "@")
	if len(mailSlice) == 2 && mailSlice[0] != "" && mailSlice[1] != "" {
		user = mailSlice[0]
		domain = mailSlice[1]
	}

	return
}
