package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eth-library-lab/dataset-dj/datastructs"
	"github.com/eth-library-lab/dataset-dj/dbutil"
	"github.com/eth-library-lab/dataset-dj/redisutil"

	"github.com/gin-gonic/gin"
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
}

func getAvailableFiles(c *gin.Context) {
	availableFiles, err := retrieveAllFiles()

	if err != nil {
		log.Fatal(err)
		c.IndentedJSON(http.StatusBadRequest, "an error occured while listing the files")
	}

	c.IndentedJSON(http.StatusOK, availableFiles)
}

// handler for inspecting the current contents of a metaArchive
func inspectArchive(c *gin.Context) {
	id := c.Param("id") // bind parameter id provided by the gin.Context object

	arch, err := dbutil.FindArchiveInDB(runfig.MongoClient, runfig.MongoCtx, id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, "archive not found")
	} else {
		c.IndentedJSON(http.StatusOK, arch)
	}

	// Check whether metaArchive exists and if so convert its list of filenames which is internally
	// saved as a set to a slice such that it can be represented in JSON
	// if arch, ok := archives[id]; ok {
	// 	c.IndentedJSON(http.StatusOK, struct {
	// 		ID    string   `json:"id"`
	// 		Files []string `json:"files"`
	// 	}{
	// 		ID:    arch.ID,
	// 		Files: arch.Files.toSlice()})
	// } else {
	// 	c.IndentedJSON(http.StatusBadRequest, "archive not found")
	// }
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
		archive, err := dbutil.FindArchiveInDB(runfig.MongoClient, runfig.MongoCtx, request.ArchiveID)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "archive not found")
		} else {
			archiveTask := request
			archiveTask.Files = archive.Files.ToSlice()

			err := redisutil.PublishArchiveTask(runfig.RdbClient, archiveTask)
			if err != nil {
				fmt.Println("error publishing archive task", err)
				c.IndentedJSON(http.StatusInternalServerError, "could not request archive download")
				return
			}
			c.IndentedJSON(http.StatusOK, archiveTask)
		}

		// if archive, ok := archives[request.ArchiveID]; ok {
		// 	downloadReq := request
		// 	downloadReq.Files = archive.Files.toSlice()
		// 	downloadFiles(downloadReq)
		// 	c.IndentedJSON(http.StatusOK, downloadReq)
		// } else {
		// 	c.IndentedJSON(http.StatusBadRequest, "archive not found")
		// 	return
		// }

	} else if request.ArchiveID != "" && len(request.Files) != 0 { // ArchiveID and Files set, Email empty
		archive, err := dbutil.FindArchiveInDB(runfig.MongoClient, runfig.MongoCtx, request.ArchiveID)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, "archive not found")
		} else {
			fileSet := datastructs.SetFromSlice(request.Files)
			unionSet := datastructs.SetUnion(fileSet, archive.Files)
			dbutil.UpdateFilesOfArchive(runfig.MongoClient, runfig.MongoCtx, request.ArchiveID, unionSet.ToSlice())
			request.Files = unionSet.ToSlice()
			c.IndentedJSON(http.StatusOK, request)
		}

		// if archive, ok := archives[request.ArchiveID]; ok {
		// 	fileSet := setFromSlice(request.Files)
		// 	archive.Files = setUnion(archive.Files, fileSet)
		// 	archives[request.ArchiveID] = archive
		// 	c.IndentedJSON(http.StatusOK, request)
		// } else {
		// 	c.IndentedJSON(http.StatusBadRequest, "archive not found")
		// 	return
		// }
	} else if request.Email != "" && len(request.Files) != 0 { // Email and Files set, ArchiveID empty

		// Create new metaArchive with random UID
		newArchive := dbutil.MetaArchive{
			ID:          generateToken(),
			Files:       datastructs.SetFromSlice(request.Files),
			TimeCreated: time.Now().String(),
			TimeUpdated: "",
			Status:      "opened",
			Source:      "local",
		}

		dbutil.AddArchiveToDB(runfig.MongoClient, runfig.MongoCtx, newArchive)

		archiveTask := request
		archiveTask.ArchiveID = newArchive.ID

		err := redisutil.PublishArchiveTask(runfig.RdbClient, archiveTask)
		if err != nil {
			fmt.Println("error publishing archive task", err)
			c.IndentedJSON(http.StatusInternalServerError, "could not request archive creation")
			return
		}
		c.IndentedJSON(http.StatusOK, newArchive)

	} else if len(request.Files) != 0 { // Files set, Email and ArchiveID empty

		// Create new metaArchive with random UID
		newArchive := dbutil.MetaArchive{
			ID:          generateToken(),
			Files:       datastructs.SetFromSlice(request.Files),
			TimeCreated: time.Now().String(),
			TimeUpdated: "",
			Status:      "opened",
			Source:      "local",
		}

		dbutil.AddArchiveToDB(runfig.MongoClient, runfig.MongoCtx, newArchive)

		c.IndentedJSON(http.StatusCreated, newArchive)
	} else {
		c.IndentedJSON(http.StatusBadRequest, "invalid request format")
	}
}

func addSourceBucket(c *gin.Context) {
	var bucket dbutil.SourceBucket

	if err := c.BindJSON(&bucket); err != nil {
		return
	}

	err := redisutil.PublishSourceBucketTask(runfig.RdbClient, bucket)
	if err != nil {
		fmt.Println("error publishing source bucket task", err)
		c.IndentedJSON(http.StatusInternalServerError, "could not request source bucket creation")
		return
	}
	c.IndentedJSON(http.StatusOK, bucket)
}

// handler for a simple healthCheck API that verifies if the service is alive / running
func healthCheck(c *gin.Context) {
	msg := "The service is running and has received the healthCheck request"
	fmt.Println(msg)
	c.IndentedJSON(http.StatusOK, msg)
}
