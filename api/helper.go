package main

import (
	"fmt"
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/eth-library/dataset-dj/util"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"time"
)

// emailIsValid if email is a valid format for a public address
// returns the parsed address and nil if valid
// or return an empty string and error if invalid
func emailIsValid(email string) (string, error) {
	e, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}
	// check that the address includes a public domain
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if emailRegex.MatchString(e.Address) != true {
		return "", fmt.Errorf("email address must be public")
	}
	return e.Address, nil
}

func createOrderForRequest(c *gin.Context, request archiveRequest) {
	sources, err := dbutil.LoadSourcesByID(runtime.MongoCtx, runtime.MongoClient, config.DbName, request.ArchiveID)
	order := dbutil.Order{
		ArchiveID: request.ArchiveID,
		Email:     request.Email,
		Date:      time.Now().String(),
		Status:    "opened",
		Sources:   sources,
	}
	res, err := dbutil.InsertOne(runtime.MongoCtx, runtime.MongoClient, config.DbName, "orders", order)
	if err != nil {
		log.Printf("Failed to create order for Archive (id: %s)", request.ArchiveID)
	}
	order.OrderID = res.InsertedID.(string)
	c.IndentedJSON(http.StatusOK, order)
}

func createArchiveForRequest(request archiveRequest) dbutil.MetaArchive {
	content, sources := dbutil.Unify(util.Mapping(request.Content, dbutil.DBToFileGroup))
	// Create new metaArchive with random UID
	newArchive := dbutil.MetaArchive{
		ID:          generateToken(),
		Content:     content,
		TimeCreated: time.Now().String(),
		TimeUpdated: "",
		Status:      "opened",
		Sources:     sources,
	}
	dbutil.AddArchiveToDB(runtime.MongoCtx, runtime.MongoClient, config.DbName, newArchive)
	return newArchive
}

func updateArchiveForRequest(c *gin.Context, request archiveRequest) {
	archive, err := dbutil.FindArchiveInDB(runtime.MongoCtx, runtime.MongoClient, config.DbName, request.ArchiveID)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "archive not found")
	} else {
		unionFileGroups, sources := dbutil.Union(archive.Content, util.Mapping(request.Content, dbutil.DBToFileGroup))
		_, err = dbutil.UpdateArchiveContent(runtime.MongoCtx, runtime.MongoClient, config.DbName, request.ArchiveID,
			unionFileGroups, sources)
		if err != nil {
			log.Printf("Failed to update content of Archive (id: %s)", request.ArchiveID)
		}
		request.Content = util.Mapping(unionFileGroups, dbutil.FileGroupToDB)
		c.IndentedJSON(http.StatusOK, request)
	}
}
