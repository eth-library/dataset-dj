package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

var (
	config        *serverConfig
	storageClient *storage.Client // client used to connect to the storage in order to read and write files
)

func main() {

	config = initServerConfig()

	ctx := context.Background()
	var err error
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storageClient.Close()

	// connect to redis instance
	rdbClient = initRedisConnection()

	router := gin.Default()
	router.GET("/files", getAvailableFilesGC)
	router.GET("/localfiles", getAvailableFilesLocal)
	router.GET("/archive/:id", inspectArchive)
	router.POST("/archive", handleArchive)
	router.GET("/check", healthCheck)
	router.Run("0.0.0.0:" + config.port) // bind to 0.0.0.0 to receive requests from outside docker container

}
