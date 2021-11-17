package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// file represents metadata about a file
type file struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Size     int32  `json:"size"`
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

func postFileList(c *gin.Context) {
	var fileNames []string

	if err := c.BindJSON(&fileNames); err != nil {
		return
	}

	err := getFiles(fileNames)
	if err != nil {
		log.Fatal(err)
	}
	err = sendNotification(fileNames)
	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusCreated, fileNames)

}
