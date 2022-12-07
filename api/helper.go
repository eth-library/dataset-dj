package main

import (
	"fmt"
	"github.com/eth-library/dataset-dj/constants"
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/eth-library/dataset-dj/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	sources, err := dbutil.LoadArchiveSources(runtime.MongoCtx, runtime.MongoClient, config.DbName, request.ArchiveID)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, fmt.Errorf("unable to load sources of archive (id: %s)",
			request.ArchiveID))
		return
	}
	order := dbutil.Order{
		OrderID:   generateID(constants.OrderIDs),
		ArchiveID: request.ArchiveID,
		Email:     request.Email,
		Date:      time.Now().String(),
		Status:    "opened",
		Sources:   sources,
	}
	// TODO | Come up with a smart way of storing the orders in the DB such that only the relevant orders are loaded
	// TODO | when a taskhandler asks for orders -> need to find orders based on provided sources
	_, err = dbutil.InsertOne(runtime.MongoCtx, runtime.MongoClient, config.DbName, constants.Orders, order)
	if err != nil {
		log.Printf("Failed to create order for Archive (id: %s)", request.ArchiveID)
	}
	c.IndentedJSON(http.StatusOK, order)
}

func createArchiveForRequest(request archiveRequest) dbutil.MetaArchive {
	content, sources := dbutil.Unify(util.Mapping(request.Content, dbutil.DBToFileGroup))
	// Create new metaArchive with random UID
	newArchive := dbutil.MetaArchive{
		ID:          generateID(constants.ArchiveIDs),
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

func generateID(col string) string {
	var ids *util.Set
	switch col {
	case constants.ArchiveIDs:
		ids = &runtime.ArchiveIDs
		break
	case constants.SourceIDs:
		ids = &runtime.SourceIDs
		break
	case constants.OrderIDs:
		ids = &runtime.OrderIDs
	}

	// Generate UUID for archives and use only the first 4 bytes
	newUID := uuid.New().String()[:8]

	// Regenerate new UUIDs as long as there are collisions
	for ok := ids.Check(newUID); ok; {
		newUID = uuid.New().String()[:8]
	}

	ids.Add(newUID)
	res, err := dbutil.UpdateIDs(runtime.MongoCtx, runtime.MongoClient, config.DbName, col, ids.ToSlice())
	if err != nil {
		log.Println(err)
	}
	log.Println(res)

	return newUID
}
