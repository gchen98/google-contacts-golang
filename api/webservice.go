package api

import (
	"github.com/gchen98/contacts_go/backend"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/people/v1"
	"log"
	"net/http"
)

//"golang.org/x/oauth2"
//"golang.org/x/oauth2"
//"encoding/json"
//"fmt"
//"os"
//"golang.org/x/oauth2/google"

func StartServer() {
	// Create a Gin router with default middleware (logger and recovery)

	r := gin.Default()

	r.GET("/ListContacts/:username", listContacts)
	r.GET("/SearchContacts/:username", searchContacts)
	// Define a simple GET endpoint
	r.POST("/AddContacts/:username", addContacts)
	r.POST("/UpdateContacts/:username", updateContacts)
	r.POST("/DeleteContacts/:username", deleteContacts)
	r.GET("/ping", func(c *gin.Context) {
		// Return JSON response
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// Start server on port 8080 (default)
	// Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	if err := r.Run("127.0.0.1:8340"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func listContacts(c *gin.Context) {
	userName := c.Param("username")
	PersonList := backend.ListContacts(userName)
	c.IndentedJSON(http.StatusOK, PersonList)
}

func searchContacts(c *gin.Context) {
	userName := c.Param("username")
	query := c.Query("query")
	PersonList := backend.SearchContacts(userName, query)
	c.IndentedJSON(http.StatusOK, PersonList)
}

// be sure to strip off id: <id> under source  and resourceName from JSON
func addContacts(c *gin.Context) {
	userName := c.Param("username")
	//userName = "garyc@caseyandgary.com"
	println("username " + userName)
	var personList []*people.Person
	if err := c.BindJSON(&personList); err != nil {
		return
	}
	personResponseList := backend.AddContacts(userName, personList)
	c.IndentedJSON(http.StatusCreated, personResponseList)
}

func updateContacts(c *gin.Context) {
	userName := c.Param("username")
	var personList []*people.Person
	if err := c.BindJSON(&personList); err != nil {
		return
	}
	personResponseMap := backend.UpdateContacts(userName, personList)
	c.IndentedJSON(http.StatusCreated, personResponseMap)
}

func deleteContacts(c *gin.Context) {
	userName := c.Param("username")
	var resourceList []string
	if err := c.BindJSON(&resourceList); err != nil {
		return
	}
	personResponseMap := backend.DeleteContacts(userName, resourceList)
	c.IndentedJSON(http.StatusCreated, personResponseMap)
}
