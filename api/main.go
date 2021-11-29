package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

var (
	// pathPrefix string = "/Users/magnuswuttke/coding/go/datadj/"
	// privateKey    []byte
	bucketName    string          // bucket name used to create bucket handlers
	storageClient *storage.Client // client used to connect to the storage in order to read and write files
)

const (
	projectID   string = "data-dj-2021"
	accessID    string = projectID + "@appspot.gserviceaccount.com" // Access ID for signing URLs
	collection  string = "data-mirror/"                             // Path to the files of the collection inside the bucket
	archStorage string = "data-archive/"                            // Path to the zipped files inside the bucket
)

func main() {
	ctx := context.Background()
	var err error

	bucketName = os.Getenv("GCLOUD_STORAGE_BUCKET") // The bucket name is set as an environment variable, see app.yaml file
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storageClient.Close()

	// connect to redis instance
	rdbClient = initRedisConnection()

	port := os.Getenv("PORT") // Retrieve the PORT env variable for usage within the google cloud
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router := gin.Default()
	router.GET("/files", getAvailableFilesGC)
	router.GET("/archive/:id", inspectArchive)
	router.POST("/archive", handleArchive)
	router.GET("/check", healthCheck)
	router.Run("0.0.0.0:" + port) // bind to 0.0.0.0 to receive requests from outside docker container

}
