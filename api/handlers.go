package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eth-library/dataset-dj/constants"

	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/gin-gonic/gin"
)

// ----------------------------------------- Authentication ------------------------------------------------

func handleCreateLink(c *gin.Context) {
	var emailRequestBody EmailRequestBody
	if err := c.BindJSON(&emailRequestBody); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "recipient email required")
		return
	}
	email := emailRequestBody.Email
	email, err := emailIsValid(email)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "email format not valid")
		return
	}

	//to prevent duplicates
	err = deleteExistingLinks(runtime.MongoCtx, runtime.MongoClient, email)
	if err != nil {
		log.Println("error deleting existing links: ", err)
	}
	//create link
	linkID := createSingleUseLink(runtime.MongoCtx, runtime.MongoClient, email)
	url := "https://" + c.Request.Host + "/key/claim/" + linkID

	//TODO: send email to recipient instead of return link
	startAPILinkEmailTask(url, email)
}

// handleValidateAPIToken provides a way to check if an Api Key is valid
func handleValidateAPIToken(c *gin.Context) {

	token, err := getTokenFromHeader(c)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	if token == "" {
		c.IndentedJSON(http.StatusUnauthorized, err.Error())
		return
	}

	collection := getColHandle("apiKeys")
	res, _ := validateAPIToken(collection, token)
	if res == false {
		c.IndentedJSON(http.StatusUnauthorized, "invalid Bearer Token")
	} else {
		c.IndentedJSON(http.StatusOK, "Authorization Bearer Token validated successfully")
	}
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

// --------------------------------------- Register new endpoints ---------------------------------------

func registerSource(c *gin.Context) {
	var request sourceRequest
	if err := c.BindJSON(&request); err != nil {
		return
	}
	source := dbutil.Source{
		ID:           generateID(constants.SourceIDs),
		Name:         request.Name,
		Organisation: request.Organisation,
	}
	dbutil.AddSourceToDB(runtime.MongoCtx, runtime.MongoClient, config.DbName, source)
	c.IndentedJSON(http.StatusOK, source.ID)
}

// registerTaskHandler and return an APIKey for it to access the "system" resources
func registerTaskHandler(c *gin.Context) {
	setupAPIToken(c, "handler")
}

// ----------------------------------------- Inspect State ----------------------------------------------

// listOrders filtered by sources specified in the orderRequest
// Need to rework the way of storing and retrieving Sources, in order to improve performance
func listOrders(c *gin.Context) {
	var request orderRequest
	if err := c.BindJSON(&request); err != nil {
		return
	}
	orders, err := dbutil.LoadOrders(runtime.MongoCtx, runtime.MongoClient, config.DbName, request.Sources)
	if err != nil {
		log.Println("ERROR retrieving requests:", err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, orders)
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

// handler for a simple healthCheck API that verifies if the service is alive / running
func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// ------------------------------------------- Archives ----------------------------------------------------

// handler for the /archive API endpoint that receives an archiveRequest. See archiveRequest for more
// information about the possible combinations that are being switched by this function.
func handleArchive(c *gin.Context) {
	var request archiveRequest
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
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
	} else if request.Email != "" && len(request.Content) != 0 { // Email and content set, ArchiveID empty
		newArchive := createArchiveForRequest(request)
		request.ArchiveID = newArchive.ID
		createOrderForRequest(c, request)
	} else if len(request.Content) != 0 { // Files set, Email and ArchiveID empty
		newArchive := createArchiveForRequest(request)
		c.IndentedJSON(http.StatusCreated, newArchive)
	} else {
		c.IndentedJSON(http.StatusBadRequest, "invalid request format")
	}
}

// --------------------------------------------- Status changes -----------------------------------------------

func updateStatus(c *gin.Context) {
	id := c.Param("id") // bind parameter id provided by the gin.Context object
	var request orderStatusRequest
	if err := c.BindJSON(&request); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	order, err := dbutil.LoadOrder(runtime.MongoCtx, runtime.MongoClient, config.DbName, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	if (order.Status == constants.Opened && request.NewStatus == constants.Processing) ||
		(order.Status == constants.Opened && request.NewStatus == constants.Rejected) ||
		(order.Status == constants.Processing && request.NewStatus == constants.Closed) {
		_, err := dbutil.UpdateOrderStatus(runtime.MongoCtx, runtime.MongoClient, config.DbName, id,
			request.NewStatus)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		_, err = dbutil.UpdateArchiveStatus(runtime.MongoCtx, runtime.MongoClient, config.DbName, order.ArchiveID,
			request.NewStatus)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(http.StatusOK, request.NewStatus)
	} else {
		c.IndentedJSON(http.StatusConflict, "Order status could not be updated")
	}
}
