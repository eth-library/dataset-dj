package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// pathPrefix string = "/Users/magnuswuttke/coding/go/datadj/"
	// privateKey    []byte
	bucket        string          // bucket name used to create bucket handlers
	storageClient *storage.Client // client used to connect to the storage in order to read and write files
	mongoClient   *mongo.Client
	mongoCtx      context.Context
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
	var cancel context.CancelFunc

	bucket = os.Getenv("GCLOUD_STORAGE_BUCKET") // The bucket name is set as an environment variable, see app.yaml file
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storageClient.Close()

	port := os.Getenv("PORT") // Retrieve the PORT env variable for usage within the google cloud
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

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
	router.GET("/files", getAvailableFilesGC)
	router.GET("/archive/:id", inspectArchive)
	router.POST("/archive", handleArchive)
	router.GET("/check", healthCheck)
	router.Run("localhost:" + port)

	// -------------------------------------------------------------------------------------------------------------------- //
	// This portion of the code is only being used when signing URLs, which is not the default behaviour of the DJ

	// This portion
	// bkt := storageClient.Bucket(bucket)
	// file := bkt.Object("key-files/data-dj-2021-9c94dd68fe31.json")
	// fileReader, err := file.NewReader(ctx)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer fileReader.Close()

	// buf := new(bytes.Buffer)
	// buf.ReadFrom(fileReader)

	// cfg, err := google.JWTConfigFromJSON(buf.Bytes())
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// privateKey = cfg.PrivateKey
}
