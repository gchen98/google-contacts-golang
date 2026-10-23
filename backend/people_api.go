package backend

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

// "fmt"
const GoogleConfigPath = "config/client_secret.json"
const ContactFieldsUpdate = "names,emailAddresses,phoneNumbers,organizations,addresses,biographies,userDefined"
const ContactFieldsRead = "names,emailAddresses,phoneNumbers,organizations,addresses,biographies,userDefined,metadata"

type JsonConfig struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
	Subject                 string `json:"subject"`
}

func InitService(username string) *people.Service {
	fileContent, err := os.ReadFile(GoogleConfigPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	var config JsonConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Fatalf("Unable to parse client secret file: %v", err)
	}
	var srv *people.Service
	ctx := context.Background() // If modifying these scopes, delete your previously saved token.json.
	conf := &jwt.Config{
		Email:      config.ClientEmail,
		PrivateKey: []byte(config.PrivateKey),
		TokenURL:   google.JWTTokenURL,
		Scopes:     []string{"https://www.google.com/m8/feeds"},
		Subject:    username,
	}
	client := conf.Client(ctx)
	srv, err = people.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create service from client secret JSON: %v", err)
	} else {
		log.Printf("Created service %v", srv)
	}
	return srv
}

func InitGmailService(username string) *gmail.Service {
	fileContent, err := os.ReadFile(GoogleConfigPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	var config JsonConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Fatalf("Unable to parse client secret file: %v", err)
	}
	var srv *gmail.Service
	ctx := context.Background() // If modifying these scopes, delete your previously saved token.json.
	conf := &jwt.Config{
		Email:      config.ClientEmail,
		PrivateKey: []byte(config.PrivateKey),
		TokenURL:   google.JWTTokenURL,
		Scopes:     []string{"https://www.googleapis.com/auth/gmail.modify"},
		Subject:    username,
	}
	client := conf.Client(ctx)
	srv, err = gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create service from client secret JSON: %v", err)
	} else {
		log.Printf("Created service %v", srv)
	}
	return srv
}

func SearchContacts(username string, searchTerm string) []*people.SearchResult {
	var SearchResultList []*people.SearchResult
	srv := InitService(username)
	SearchResponse, err := srv.People.SearchContacts().
		Query(searchTerm).
		ReadMask(ContactFieldsRead).
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve Katie. %v", err)
	} else {
		SearchResultList = SearchResponse.Results

		//j, err := json.Marshal(resp)
		//if err != nil {
		//	log.Fatalf("Unable to parse response to Katie. %v", err)
		//} else {
		//	log.Printf(string(j))
		//}
	}
	return SearchResultList

}
func ListLabels(username string) []*gmail.Label {
	log.Printf("Listing labels for %v", username)
	srv := InitGmailService(username)
	var LabelList []*gmail.Label
	ListLabelsResponse, Err := srv.Users.Labels.List(username).Do()
	if Err != nil {
		log.Printf("Unable to list labels for %v. %v", username, Err.Error())
	} else {
		for _, Label := range ListLabelsResponse.Labels {

			if strings.HasPrefix(Label.Id, "Label_") || Label.Name == "SENT" {
				log.Printf("Label ID %v with name: %v", Label.Id, Label.Name)
				LabelList = append(LabelList, Label)
			}

		}
	}
	return LabelList
}

func ListAndModifyGmail(username string, label string) {

	srv := InitGmailService(username)
	RemoveIds := []string{"INBOX"}
	FilterLabels := []string{"INBOX"}
	FilterLabels = append(FilterLabels, label)
	log.Printf("Removing INBOX tax for messages with label %v", FilterLabels)
	UsersMessagesListCall := srv.Users.Messages.List(username).LabelIds(FilterLabels...).MaxResults(500)
	for {
		ListMessagesResponse, Err := UsersMessagesListCall.Do()
		if Err != nil {
			log.Printf("Unable to list Gmail for %v. %v", username, Err.Error())
			break
		}
		if len(ListMessagesResponse.Messages) == 0 {
			log.Printf("Nothing to do there are no matching messages")
			break
		}
		// build the ID list
		ModifyMessageIdList := []string{}
		for _, Message := range ListMessagesResponse.Messages {
			//log.Printf("Adding Message ID: %v", Message.Id)
			ModifyMessageIdList = append(ModifyMessageIdList, Message.Id)
		}
		// send the ID list as a batchmodify
		BatchModifyMessagesRequest := &gmail.BatchModifyMessagesRequest{
			Ids:            ModifyMessageIdList,
			RemoveLabelIds: RemoveIds,
		}
		log.Printf("Will remove %v tags ", len(ModifyMessageIdList))
		Err2 := srv.Users.Messages.BatchModify(username, BatchModifyMessagesRequest).Do()
		if Err2 != nil {
			log.Printf("Unable to modify Gmail for %v. %v", username, Err.Error())
		} else {
			log.Printf("Successfully modified Gmail for %v", username)
		}

		NextPageToken := ListMessagesResponse.NextPageToken
		log.Printf("Next page Token %s", NextPageToken)
		if len(NextPageToken) > 0 {
			UsersMessagesListCall.PageToken(NextPageToken)
		} else {
			log.Println("No NextPageToken")
			break
		}
		time.Sleep(1 * time.Second)
		//break
	}
	time.Sleep(1 * time.Second)
}

func ListGmail(username string, label string) []string {
	//LabelList := []string{"INBOX"}
	//LabelList = append(LabelList, label)
	//Labels := strings.Join(LabelList, ",")

	log.Printf("Listing Gmail for %v with labels %v", username, label)
	srv := InitGmailService(username)
	//log.Printf("2 Listing Gmail for %v", username)
	//UsersMessagesListCall := srv.Users.Messages.
	//	List(username).
	//	MaxResults(500)
	UsersMessagesListCall := srv.Users.Messages.List(username).LabelIds("INBOX").LabelIds(label).MaxResults(500)
	ModifyMessageIdList := []string{}
	for {
		ListMessagesResponse, Err := UsersMessagesListCall.Do()
		if Err != nil {
			log.Printf("Unable to list Gmail for %v. %v", username, Err.Error())
			break
		}
		for _, Message := range ListMessagesResponse.Messages {
			//log.Printf("Adding Message ID: %v", Message.Id)
			ModifyMessageIdList = append(ModifyMessageIdList, Message.Id)
		}
		NextPageToken := ListMessagesResponse.NextPageToken
		log.Printf("Next page Token %s", NextPageToken)
		if len(NextPageToken) > 0 {
			UsersMessagesListCall.PageToken(NextPageToken)
		} else {
			log.Println("No NextPageToken")
			break
		}
		break
	}
	log.Printf("Modify message list has size %v", len(ModifyMessageIdList))
	return ModifyMessageIdList
}

func ListContacts(username string) []*people.Person {
	srv := InitService(username)
	const pageSize = 500
	PeopleConnectionsListCall := srv.People.Connections.
		List("people/me").
		PageSize(pageSize).
		PersonFields(ContactFieldsRead)
	var PersonList []*people.Person
	for {
		ListConnectionsResponse, err := PeopleConnectionsListCall.Do()
		if err != nil {
			log.Printf("Unable to retrieve people. %v", err)
			break
		}
		if len(ListConnectionsResponse.Connections) > 0 {
			log.Printf("List %v connection names:\n", pageSize)
			for _, Person := range ListConnectionsResponse.Connections {
				//names := Person.Names
				//if len(names) > 0 {
				//name := names[0].DisplayName
				//fmt.Printf("%s\n", name)
				//}
				PersonList = append(PersonList, Person)
			}
			//json, err := r.MarshalJSON()
			//if err != nil {
			//	log.Fatalf("Unable to marshal json. %v", err)
			//} else {
			//	fmt.Println(string(json))
			//}
			NextPageToken := ListConnectionsResponse.NextPageToken
			log.Printf("Next page Token %s", NextPageToken)
			if len(NextPageToken) > 0 {
				PeopleConnectionsListCall.PageToken(NextPageToken)
			} else {
				log.Println("No NextPageToken")
				break
			}
		} else {
			log.Println("No connections found.")
			break
		}
	}
	return PersonList
}

func DeleteContacts(username string, deleteList []string) string {
	srv := InitService(username)
	batchDeleteContactsRequest := &people.BatchDeleteContactsRequest{
		ResourceNames: deleteList,
	}
	_, err := srv.People.BatchDeleteContacts(batchDeleteContactsRequest).Do()
	if err != nil {
		log.Printf("Unable to delete contacts. %v", err)
		return err.Error()
	} else {
		return "Successfully deleted"
	}
}
func AddContacts(username string, personList []*people.Person) []*people.PersonResponse {
	srv := InitService(username)
	var contacts []*people.ContactToCreate
	for _, Person := range personList {
		contactToCreate := &people.ContactToCreate{
			ContactPerson: Person,
		}
		contacts = append(contacts, contactToCreate)
	}
	contactsToCreate := &people.BatchCreateContactsRequest{
		Contacts: contacts,
		ReadMask: ContactFieldsRead,
	}
	resp, err := srv.People.BatchCreateContacts(contactsToCreate).Do()
	if err != nil {
		log.Printf("Unable to create batch contacts. %v", err)
		var empty []*people.PersonResponse
		return empty
	} else {
		r, _ := resp.MarshalJSON()
		log.Printf("Created batch contacts %v", string(r))
		return resp.CreatedPeople
		//return "Successfully added contacts"
	}
}

func UpdateContacts(username string, personList []*people.Person) map[string]people.PersonResponse {
	srv := InitService(username)
	//var contacts []*people.ContactToCreate
	contactsMap := make(map[string]people.Person)

	for _, Person := range personList {
		//contactToCreate := &people.ContactToCreate{
		//	ContactPerson: Person,
		//}
		//contacts = append(contacts, contactToCreate)
		resourceName := Person.ResourceName
		contactsMap[resourceName] = *Person
	}

	contactsToUpdate := &people.BatchUpdateContactsRequest{
		Contacts:   contactsMap,
		ReadMask:   ContactFieldsRead,
		UpdateMask: ContactFieldsUpdate,
	}
	resp, err := srv.People.BatchUpdateContacts(contactsToUpdate).Do()

	if err != nil {
		log.Printf("Unable to update batch contacts. %v", err)
		return make(map[string]people.PersonResponse)
	} else {
		r, _ := resp.MarshalJSON()
		log.Printf("Updated batch contacts %v", string(r))
		//return "Successfully updated contacts"
		return resp.UpdateResult
	}
}
