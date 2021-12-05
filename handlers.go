package main

import (
	"fmt"
	"log"
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

		filterMap := buildReqAttributesMap(buildReqAttributesList(searchReq.Filter, []string{}))
		filter := getAzureFieldsMap(filterMap)

		usersFound := findUsers(filter, searchType)

		if len(usersFound) > 0 {
			var entries []*ldap.Entry

			for _, u := range usersFound {
				var mail []string
				userPrincipalName, ok := u["userPrincipalName"].(string)
				if ok {
					mail = strings.Split(userPrincipalName, "@")
				}

				if len(mail) > 0 {
					baseDN := getDN(mail[1])
					var entry ldap.Entry

					entry.DN = fmt.Sprintf("cn=%s,%s", mail[0], baseDN)
					entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: "username", Values: []string{mail[0]}})
					entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: "mail", Values: []string{userPrincipalName}})
					entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"person"}})
					for k, value := range u {
						ldapField := getLDAPField(k)
						if ldapField != "" {
							switch v := value.(type) {
							case string:
								entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapField, Values: []string{v}})
							default:
								entry.Attributes = append(entry.Attributes, &ldap.EntryAttribute{Name: ldapField, Values: []string{fmt.Sprint(v)}})
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

func getAzureFieldsMap(filter map[string]string) map[string]string {
	azureFields := make(map[string]string)
	for k := range filter {
		value, ok := appConfig.mapFields[k]
		if ok {
			azureFields[value] = filter[k]
		}
	}

	return azureFields
}

func findUsers(filter map[string]string, searchType byte) (usersFound []map[string]interface{}) {
	users, err := getUsersMap()
	if err == nil {
		for _, v := range users {
			found := false
			for k, f := range filter {
				/*fieldValue, ok := v[k]
				if ok {
					found = f == fieldValue
					if found && searchType != '&' {
						break
					}
				}
				if !found && searchType == '&' {
					break
				}*/
				fieldValue, ok := v[k]
				if ok {
					found = f == fieldValue
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
			}
			if found {
				usersFound = append(usersFound, v)
			}
		}
	}

	return
}

func getLDAPField(azureField string) string {
	for k, v := range appConfig.mapFields {
		if v == azureField {
			return k
		}
	}

	return ""
}

/*func (h ldapHandler) SearchFirst(boundDN string, searchReq ldap.SearchRequest, conn net.Conn) (ldap.ServerSearchResult, error) {
	if boundDN != "" {
		util.AddToLog("search", fmt.Sprintf("%+v", searchReq))

		searchType := searchReq.Filter[1]           // get filter type
		if searchType != '|' && searchType != '&' { // check filter type | &
			searchType = '|'
		}

		filters := buildReqAttributesList(searchReq.Filter, []string{})
		filterMap := buildReqAttributesMap(filters)

		empId, ok := filterMap["employeeid"]
		if ok {
			response, err := getUser(empId)
			mail := strings.Split(response.UserPrincipalName, "@")
			if err == nil && len(mail) > 0 {
				baseDN := getDN(mail[1])

				entries := []*ldap.Entry{
					&ldap.Entry{fmt.Sprintf("cn=%s,%s", mail[0], baseDN), []*ldap.EntryAttribute{
						&ldap.EntryAttribute{"username", []string{mail[0]}},
						&ldap.EntryAttribute{"mail", []string{response.UserPrincipalName}},
						&ldap.EntryAttribute{"employeeid", []string{response.EmployeeID}},
						&ldap.EntryAttribute{"objectClass", []string{"person"}},
					}}}

				return ldap.ServerSearchResult{entries, []string{}, []ldap.Control{}, ldap.LDAPResultSuccess}, nil
			} else {
				util.AddToLog("search", fmt.Sprintf("employeeId=%s not found in Azure", empId))
			}
		} else {
			util.AddToLog("search", "employeedId filter not found")
		}
	} else {
		util.AddToLog("search", "not bound")
	}

	return ldap.ServerSearchResult{nil, []string{}, []ldap.Control{}, ldap.LDAPResultOperationsError}, nil
}*/

func (h ldapHandler) Unbind(boundDN string, conn net.Conn) (ldap.LDAPResultCode, error) {
	log.Printf("Unbind: %+v", boundDN)

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

func buildReqAttributesMap(filter []string) (filterMap map[string]string) {
	filterMap = make(map[string]string)
	for _, v := range filter {
		f := strings.Split(v, "=")
		if len(f) > 1 {
			filterMap[strings.ToLower(f[0])] = f[1]
		}
	}

	return
}

func getDN(domain string) (baseDN string) {
	domainParts := strings.Split(domain, ".")
	for _, v := range domainParts {
		baseDN += fmt.Sprintf("dc=%s,", v)
	}
	baseDN = strings.TrimSuffix(baseDN, ",")

	return
}
