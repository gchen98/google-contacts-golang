package backend

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"google.golang.org/api/people/v1"
)

func TestUpdateContacts(t *testing.T) {
	t.Skip("Not testing Contacts")
	fileContent, err := os.ReadFile("update_contacts.json")
	if err != nil {
		log.Fatalf("Unable to parse update_contacts.json: %v", err)
	}
	var personList []*people.Person
	log.Printf("Unmarshalling file: %s", fileContent)
	err = json.Unmarshal(fileContent, &personList)
	if err != nil {
		log.Fatalf("Unable to Unmarshal JSON to add list: %v", err)
	} else {
		log.Printf("Unmarshalled personList: %v", personList)
	}
	UpdateContacts("garyc@caseyandgary.com", personList)
}

func TestAddContacts(t *testing.T) {
	t.Skip("Not testing Contacts")
	fileContent, err := os.ReadFile("add_contacts.json")
	if err != nil {
		log.Fatalf("Unable to parse add_contacts.json: %v", err)
	}
	var personList []*people.Person
	log.Printf("Unmarshalling file: %s", fileContent)
	err = json.Unmarshal(fileContent, &personList)
	if err != nil {
		log.Fatalf("Unable to Unmarshal JSON to add list: %v", err)
	} else {
		log.Printf("Unmarshalled personList: %v", personList)
	}
	AddContacts("garyc@caseyandgary.com", personList)
}

func TestDeleteContacts(t *testing.T) {
	t.Skip("Not testing Contacts")
	fileContent, err := os.ReadFile("delete.json")
	if err != nil {
		log.Fatalf("Unable to parse delete.json: %v", err)
	}
	var deleteList []string
	log.Printf("Unmarshalling file: %s", fileContent)
	err = json.Unmarshal(fileContent, &deleteList)
	if err != nil {
		log.Fatalf("Unable to Unmarshal JSON to delete list: %v", err)
	}
	log.Printf("Delete list: %v", deleteList)
	message := DeleteContacts("garyc@caseyandgary.com", deleteList)
	log.Printf("Status %v", message)

}
