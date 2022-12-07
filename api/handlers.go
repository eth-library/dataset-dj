package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/eth-library/dataset-dj/redisutil"

	"github.com/gin-gonic/gin"
)

// listArchives retrieves all MetaArchive headers from the database
func listOrders(c *gin.Context) {
	requests, err := dbutil.LoadOrders(runtime.MongoCtx, runtime.MongoClient, config.DbName)
	if err != nil {
		log.Println("ERROR retrieving requests:", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, requests)
}

// claimKey for API usage with "service" permissions
func claimKey(c *gin.Context) {
	linkID := c.Param("id")
	linkValid, err := validateTokenLink(runtime.MongoCtx, runtime.MongoClient, linkID)
	if err != nil {
		log.Println("ERROR validating Token Link:", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}
	if linkValid != true {
		c.IndentedJSON(http.StatusBadRequest, "invalid link")
		return
	}
	setupAPIToken(c, "service")
}

// registerTaskHandler and return an APIKey for it to access the "system" resources
func registerTaskHandler(c *gin.Context) {
	setupAPIToken(c, "system")
}

// inspectArchive to receive its current contents
func inspectArchive(c *gin.Context) {
	id := c.Param("id") // bind parameter id provided by the gin.Context object

	arch, err := dbutil.FindArchiveInDB(runtime.MongoCtx, runtime.MongoClient, config.DbName, id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "archive not found")
	} else {
		c.IndentedJSON(http.StatusOK, arch)
	}
}

// handler for the /archive API endpoint that receives an archiveRequest. See archiveRequest for more
// information about the possible combinations that are being switched by this function.
func handleArchive(c *gin.Context) {
	var request archiveRequest
	if err := c.BindJSON(&request); err != nil {
		return
	}

	// validate email format
	if request.Email != "" {
		if _, err := emailIsValid(request.Email); err != nil {
			c.IndentedJSON(http.StatusBadRequest, "invalid email address")
			return
		}
	}

	if request.Email != "" && request.ArchiveID != "" { // Email and ArchiveID set
		if !runtime.ArchiveIDs.Check(request.ArchiveID) {
			c.IndentedJSON(http.StatusBadRequest, fmt.Sprintf("Archive (id: %s) does not exist",
				request.ArchiveID))
		}
		createOrderForRequest(c, request)
	} else if request.ArchiveID != "" && len(request.Content) != 0 { // ArchiveID and Files set, Email empty
		updateArchiveForRequest(c, request)
	} else if request.Email != "" && len(request.Content) != 0 { // Email and Files set, ArchiveID empty
		createArchiveForRequest(request)
		createOrderForRequest(c, request)
	} else if len(request.Content) != 0 { // Files set, Email and ArchiveID empty
		newArchive := createArchiveForRequest(request)
		c.IndentedJSON(http.StatusCreated, newArchive)
	} else {
		c.IndentedJSON(http.StatusBadRequest, "invalid request format")
	}
}

func getAvailableFiles(c *gin.Context) {
	availableFiles, err := retrieveAllFiles()

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "an error occured while listing the files")
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, availableFiles)
}

func addSourceBucket(c *gin.Context) {
	var bucket dbutil.SourceBucket

	if err := c.BindJSON(&bucket); err != nil {
		fmt.Println(err)
		return
	}

	err := redisutil.PublishSourceBucketTask(runtime.RdbClient, bucket)
	if err != nil {
		fmt.Println("error publishing source bucket task", err)
		c.IndentedJSON(http.StatusInternalServerError, "could not request source bucket creation")
		return
	}
	c.IndentedJSON(http.StatusOK, bucket)
}

// handler for a simple healthCheck API that verifies if the service is alive / running
func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
