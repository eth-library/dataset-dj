package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	config        *serverConfig
	storageClient *storage.Client // client used to connect to the storage in order to read and write files
	mongoClient   *mongo.Client
	mongoCtx      context.Context
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

	var cancel context.CancelFunc

	// Get Client, Context, CalcelFunc and
	// err from connect method.
	mongoClient, mongoCtx, cancel, err = connectToMDB("mongodb+srv://data-dj:LibLab123@archive-cluster.jzmhu.mongodb.net/data-dj-main?retryWrites=true&w=majority")
	if err != nil {
		panic(err)
	}

	// Release resource when the main
	// function is returned.
	defer closeMDB(mongoClient, mongoCtx, cancel)

	// Ping mongoDB with Ping method
	err = pingMDB(mongoClient, ctx)
	if err != nil {
		fmt.Println(err)
	}

	// Load the list of already used archiveIDs when redeploying
	archiveIDs, err = loadArchiveIDs()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/files", getAvailableFiles)
	router.GET("/archive/:id", inspectArchive)
	router.POST("/archive", handleArchive)
	router.GET("/check", healthCheck)
	router.Run("0.0.0.0:" + config.port) // bind to 0.0.0.0 to receive requests from outside docker container

}
