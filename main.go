package main

import (
	"github.com/gin-gonic/gin"
)

var pathPrefix string = "/Users/magnuswuttke/coding/go/datadj/"
var collection string = pathPrefix + "collection/"
var storage string = pathPrefix + "storage/"
var archiveName = "archive.zip"

func main() {
	router := gin.Default()
	router.GET("/filelist", getAvailableFiles)
	router.POST("/getFiles", postFileList)
	router.Run("localhost:8080")
}
