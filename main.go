package main

import (
	"github.com/gin-gonic/gin"
)

var pathPrefix string = "/Users/magnuswuttke/coding/go/datadj/"
var collection string = pathPrefix + "collection/"
var storage string = pathPrefix + "storage/"

func main() {
	router := gin.Default()
	router.GET("/files", getAvailableFiles)
	router.GET("/archive/:id", inspectArchive)
	router.POST("/archive", handleArchive)
	router.Run("localhost:8080")
}
