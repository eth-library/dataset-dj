package main

import (
	"context"
	"fmt"
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/eth-library/dataset-dj/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"time"

	"cloud.google.com/go/storage"
	conf "github.com/eth-library/dataset-dj/configuration"
)

// simple "Database" for the metaArchives
// var archives map[string]metaArchive = make(map[string]metaArchive)

// File represents metadata about a file, not used so far
type File struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Size     int32  `json:"size"`
}

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
	order := dbutil.Order{
		ArchiveID: request.ArchiveID,
		Email:     request.Email,
		Date:      time.Now().String(),
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

// retrieve file names from local storage, from cloud storage and also from storages that
// are connected via API. This function acts as layer of abstraction such that the function
// calls in handlers.go don't need to be modified.
func retrieveAllFiles() ([]string, error) {
	override := true
	var allAvailableFiles []string
	localFiles, err := retrieveFilesLocal(config.SourceLocalDir)
	if err != nil {
		return nil, err
	}
	allAvailableFiles = append(allAvailableFiles, localFiles...)

	if len(runtime.SourceBucketList) > 0 || override {
		cloudFiles, err := retrieveFilesCloud(runtime.StorageClient, config)
		if err != nil {
			return allAvailableFiles, err
		}
		allAvailableFiles = append(allAvailableFiles, cloudFiles...)
	}

	apiFiles, err := retriveFilesAPI()
	if err != nil {
		return allAvailableFiles, err
	}
	allAvailableFiles = append(allAvailableFiles, apiFiles...)
	return allAvailableFiles, nil
}

// retrieve file names from local storage (a directory that may be accessed directly)
func retrieveFilesLocal(localSourceDir string) ([]string, error) {
	if localSourceDir == "" {
		return []string{}, nil
	}
	return listFileDir(localSourceDir)
}

// retrieve file names from cloud storage (google cloud bucket)
func retrieveFilesCloud(client *storage.Client, config *conf.ServerConfig) ([]string, error) {
	ctx := context.Background()
	var cloudFiles []string

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// get bucket handler and obtain an iterator over all objects returned by query

	bucket := client.Bucket(config.SourceBucketName)

	it := bucket.Objects(ctx, &storage.Query{
		Prefix:    config.SourceBucketPrefix,
		Delimiter: "/",
	})

	// Loop over all objects returned by the query
	for {
		attrs, err := it.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("an error occured while retrieving a file from the cloud storage: %s", err)
		}

		if attrs.Name == config.SourceBucketPrefix { // make sure the directory is not listed as available file
			continue
		}
		cloudFiles = append(cloudFiles, "cloud/"+attrs.Name)
	}
	return cloudFiles, nil
}

// retrieve file names from storages connected via API (not defined yet)
func retriveFilesAPI() ([]string, error) {
	return []string{}, nil
}

// list names of files in the given directory
func listFileDir(dirPath string) ([]string, error) {

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var filenames []string

	for _, file := range files {
		filenames = append(filenames, "local/"+file.Name())
		//print filename and if its a direcory
		// fmt.Println(file.Name(), file.IsDir())
	}

	return filenames, nil
}
