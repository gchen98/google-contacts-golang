package api

import (
	"log"
	"net/http"

	"github.com/gchen98/contacts_go/backend"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/people/v1"
)

//"golang.org/x/oauth2"
//"golang.org/x/oauth2"
//"encoding/json"
//"fmt"
//"os"
//"golang.org/x/oauth2/google"

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func StartServer() {
	// Create a Gin router with default middleware (logger and recovery)

	r := gin.Default()

	r.GET("/ListContacts/:username", listContacts)
	r.GET("/SearchContacts/:username", searchContacts)
	//r.GET("/albums/:id", getAlbumByID)
	//r.GET("/albums", getAlbums)
	// Define a simple GET endpoint
	r.POST("/AddContacts/:username", addContacts)
	r.POST("/UpdateContacts/:username", updateContacts)
	r.POST("/DeleteContacts/:username", deleteContacts)
	//r.POST("/albums", postAlbums)
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

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)

}

func postAlbums(c *gin.Context) {
	var newAlbum album
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func getAlbumByID(c *gin.Context) {
	id := c.Param("id")
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
