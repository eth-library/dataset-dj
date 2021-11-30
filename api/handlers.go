package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

// archiveRequest is the main data structure that is being received by the API when information or
// modifications about archives are requested. Email simply is an email as string, ArchiveID is the UID of
// a metaArchive as string and Files is a list of fileNames as strings. Possible combinations:
// 1. Email and ArchiveID set, Files empty -> Retrieve metaArchive with ArchiveID, create the zip archive
// 	  and send download link to Email
// 2. ArchiveID and Files set, Email empty -> Add fileNames in Files to metaArchive with ArchiveID
// 3. Email and Files set, ArchiveID empty -> Create new metaArchive containing the fileNames from Files,
//	  immediatly retrieve the files from the collection and create the zip archive and send the download
//    link to Email
// 4. Files set, Email and ArchiveID empty -> Create new metaArchive from the fileNames in Files
//
// The function handleArchive() implements the logic to identify the different cases and to act accordingly

type archiveRequest struct {
	Email     string   `json:"email"`
	ArchiveID string   `json:"archiveID"`
	Files     []string `json:"files"`
	Source    string   `json:"source"`
}

//getAvailableFilesLocal gets a list of all available filenames inside a local directory as JSON
// mostly intended as a debugging/development aid rather than a end-user endpoint
func getAvailableFilesLocal(c *gin.Context) {
	var availableFiles []string
	//TO DO return bad request of no source local dir
	availableFiles, err := listFileDir(config.sourceLocalDir)

	if err != nil {
		log.Fatal(err)
		c.IndentedJSON(http.StatusBadRequest, "directory not found")
	}

	c.IndentedJSON(http.StatusOK, availableFiles)
}

// get a list of all available files inside the collection defined by prefix on google cloud
func getAvailableFilesGC(c *gin.Context) {
	ctx := context.Background()
	availableFiles := []string{}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// get bucket handler and obtain an iterator over all objects returned by query

	bucket := storageClient.Bucket(config.sourceBucketName)

	it := bucket.Objects(ctx, &storage.Query{
		Prefix: config.sourceBucketPrefix,
	})

	// Loop over all objects returned by the query
	for {
		attrs, err := it.Next()
		if err != nil {
			log.Fatal(err)
			c.IndentedJSON(http.StatusBadRequest, "An error occured while listing the files")
		}

		if err == iterator.Done {
			break
		}

		if attrs.Name == config.sourceBucketPrefix { // make sure the directory is not listed as available file
			continue
		}
		availableFiles = append(availableFiles, attrs.Name)
	}
	c.IndentedJSON(http.StatusOK, availableFiles)
}

// handler for inspecting the current contents of a metaArchive
func inspectArchive(c *gin.Context) {
	id := c.Param("id") // bind parameter id provided by the gin.Context object

	// Check whether metaArchive exists and if so convert its list of filenames which is internally
	// saved as a set to a slice such that it can be represented in JSON
	if arch, ok := archives[id]; ok {
		c.IndentedJSON(http.StatusOK, struct {
			ID    string   `json:"id"`
			Files []string `json:"files"`
		}{
			ID:    arch.ID,
			Files: arch.Files.toSlice()})
	} else {
		c.IndentedJSON(http.StatusBadRequest, "archive not found")
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
		if !emailIsValid(request.Email) {
			c.IndentedJSON(http.StatusBadRequest, "invalid email address")
			return
		}
	}

	if request.Email != "" && request.ArchiveID != "" { // Email and ArchiveID set
		fmt.Println("Email and ArchiveID set")
		if archive, ok := archives[request.ArchiveID]; ok { // TO DO replace in-memory map lookup with db query
			downloadReq := request
			downloadReq.Files = archive.Files.toSlice()

			fmt.Println("TO DO write new metaArchive to db")
			// TO DO write new metaArchive to db

			c.IndentedJSON(http.StatusOK, downloadReq)
		} else {
			c.IndentedJSON(http.StatusBadRequest, "archive not found")
			return
		}
	} else if request.ArchiveID != "" && len(request.Files) != 0 { // ArchiveID and Files set, Email empty
		if archive, ok := archives[request.ArchiveID]; ok {
			fileSet := setFromSlice(request.Files)
			archive.Files = setUnion(archive.Files, fileSet)

			fmt.Println("TO DO update metaArchive in db")
			// TO DO update metaArchive in db

			c.IndentedJSON(http.StatusOK, request)
		} else {
			c.IndentedJSON(http.StatusBadRequest, "archive not found")
			return
		}
	} else if request.Email != "" && len(request.Files) != 0 { // Email and Files set, ArchiveID empty

		if request.Source == "" {
			request.Source = "cloud"
		}
		// Create new metaArchive with random UID
		newArchive := metaArchive{
			ID:          generateToken(),
			Files:       setFromSlice(request.Files),
			TimeCreated: time.Now().String(),
			TimeUpdated: "",
			Status:      "opened",
			Source:      "local",
		}

		fmt.Println("TO DO write new metaArchive to db")
		//TO DO store newMetaArchive in DB

		archiveTask := request
		archiveTask.ArchiveID = newArchive.ID

		err := publishArchiveTask(archiveTask)
		if err != nil {
			fmt.Println("error publishing archive task", err)
			c.IndentedJSON(http.StatusInternalServerError, "could not request archive creation")
			return
		}
		c.IndentedJSON(http.StatusOK, newArchive)

	} else if len(request.Files) != 0 { // Files set, Email and ArchiveID empty

		// Create new metaArchive with random UID
		newArchive := metaArchive{
			ID:          generateToken(),
			Files:       setFromSlice(request.Files),
			TimeCreated: time.Now().String(),
			TimeUpdated: "",
			Status:      "opened",
			Source:      "local",
		}

		fmt.Println("TO DO write new metaArchive to db")
		//TO DO store newMetaArchive in DB

		c.IndentedJSON(http.StatusCreated, newArchive)
	} else {
		c.IndentedJSON(http.StatusBadRequest, "invalid request format")
	}
}

// handler for a simple healthCheck API that verifies if the service is alive / running
func healthCheck(c *gin.Context) {
	msg := "The service is running and has received the healthCheck request"
	fmt.Println(msg)
	c.IndentedJSON(http.StatusOK, msg)
}
