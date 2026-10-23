package backend

import (
	"fmt"
	"log"
	"math"

	"github.com/go-ldap/ldap/v3"
)

func TestLdap() {
	//var filter = "(&(objectClass=mozillaAbPersonAlpha)(|(ou=frank)(ou=alice))(cn=Yun-Hua Lee))"
	//log.Printf(filter)

	conn, err := ldap.DialURL("ldap://lupine:389")
	if err != nil {
		log.Fatalf("Unable to connect to LDAP: %v", err)
	}
	defer conn.Close()

	var pageSize uint32 = 10
	searchBase := "ou=Public,dc=pioneerind,dc=com"
	filter := "(&(objectClass=mozillaAbPersonAlpha)(|(ou=frank)))"
	pagingControl := ldap.NewControlPaging(pageSize)
	attributes := []string{"dn", "title", "ou"}
	controls := []ldap.Control{pagingControl}
	err = conn.Bind("cn=admin,dc=pioneerind,dc=com", "password")
	if err != nil {
		log.Fatalf("Unable to bind LDAP: %v", err)
	}

	for {
		//controls := []ldap.Control{}

		request := ldap.NewSearchRequest(
			searchBase,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			math.MaxInt32,
			0,
			false,
			filter,
			attributes,
			controls)

		response, err := conn.Search(request)
		if err != nil {
			log.Fatalf("Unable to search for DNs: %v", err)
		} else {
			for _, entry := range response.Entries {
				fmt.Println(entry)
			}
		}

		// In order to prepare the next request, we check if the response
		// contains another ControlPaging object and a not-empty cookie and
		// copy that cookie into our pagingControl object:
		updatedControl := ldap.FindControl(response.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pagingControl.SetCookie(ctrl.Cookie)
			log.Printf("Getting another page with %v", ctrl.Cookie)
			continue
		}
		// If no new paging information is available or the cookie is empty, we
		// are done with the pagination.
		break
	}
}
