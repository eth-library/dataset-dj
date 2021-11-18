package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var archives map[string]metaArchive = make(map[string]metaArchive)

type archiveRequest struct {
	Email     string   `json:"email"`
	ArchiveID string   `json:"archiveID"`
	Files     []string `json:"files"`
}

// file represents metadata about a file
type file struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Size     int32  `json:"size"`
}

type metaArchive struct {
	ID    string   `json:"id"`
	Files []string `json:"files"`
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
		filenames = append(filenames, file.Name())
		//print filename and if its a direcory
		// fmt.Println(file.Name(), file.IsDir())
	}

	return filenames, nil
}

// getAvailableFiles responds with the list of all available files as JSON.
func getAvailableFiles(c *gin.Context) {
	var availableFiles []string

	dirPath := collection
	availableFiles, err := listFileDir(dirPath)

	if err != nil {
		log.Fatal(err)
		c.IndentedJSON(http.StatusBadRequest, "directory not found")
	}

	c.IndentedJSON(http.StatusOK, availableFiles)
}

func inspectArchive(c *gin.Context) {
	id := c.Param("id")

	if arch, ok := archives[id]; ok {
		c.IndentedJSON(http.StatusOK, arch)
	} else {
		c.IndentedJSON(http.StatusBadRequest, "archive not found")
	}
}

func handleArchive(c *gin.Context) {
	var request archiveRequest

	if err := c.BindJSON(&request); err != nil {
		return
	}

	if request.Email != "" && request.ArchiveID != "" {
		if archive, ok := archives[request.ArchiveID]; ok {
			downloadReq := request
			downloadReq.Files = archive.Files
			downloadFiles(downloadReq)
			c.IndentedJSON(http.StatusOK, downloadReq)
		} else {
			c.IndentedJSON(http.StatusBadRequest, "archive not found")
			return
		}
	} else if request.ArchiveID != "" && len(request.Files) != 0 {
		if archive, ok := archives[request.ArchiveID]; ok {
			archive.Files = append(archive.Files, request.Files...)
			archives[request.ArchiveID] = archive
			c.IndentedJSON(http.StatusOK, request)
		} else {
			c.IndentedJSON(http.StatusBadRequest, "archive not found")
			return
		}
	} else if request.Email != "" && len(request.Files) != 0 {
		newArch := metaArchive{ID: generateToken(), Files: request.Files}
		archives[newArch.ID] = newArch
		downloadReq := request
		downloadReq.ArchiveID = newArch.ID
		downloadFiles(downloadReq)
		c.IndentedJSON(http.StatusOK, downloadReq)

	} else if len(request.Files) != 0 {
		archive := metaArchive{ID: generateToken(), Files: request.Files}
		archives[archive.ID] = archive
		request.ArchiveID = archive.ID
		c.IndentedJSON(http.StatusCreated, request)
	} else {
		c.IndentedJSON(http.StatusBadRequest, "invalid request format")
	}
}

func downloadFiles(request archiveRequest) {
	err := getFiles(request)
	if err != nil {
		log.Fatal(err)
	}
	err = sendNotification(request)
	if err != nil {
		log.Fatal(err)
	}
}
