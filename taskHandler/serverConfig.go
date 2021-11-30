package main

import (
	"fmt"
	"log"
	"os"
)

// serverConfig holds all of the deployment specific environment variables and settings
// it gets initialised by a function and should be globally accessible
type serverConfig struct {
	projectID           string // google cloud project-id that contains bucket resources
	archiveBaseURL      string // url where zip files will be accessible from
	sourceBucketName    string // name of bucket where provided files are stored
	sourceBucketPrefix  string // only returns files within this directory in the bucket
	archiveBucketName   string // name of bucket where archive files will be written to
	archiveBucketPrefix string // path to directory in the bucket to write archive files to
	sourceLocalDir      string // path to directory on local machine with available files
	archiveLocalDir     string // path to directory on local machine to write archive files
	port                string // port to start api listening on
	mode                string // mode that gin is running in (use as mode for entire application)
}

//initServerConfig initiliases the serverConfig struct with all of the environment variables
func initServerConfig() *serverConfig {
	cfg := serverConfig{
		projectID:           os.Getenv("PROJECT_ID"), //for example: "data-dj-2021
		archiveBaseURL:      os.Getenv("ARCHIVE_BASE_URL"),
		sourceBucketName:    os.Getenv("SOURCE_BUCKET_NAME"),    //for example: "data-dj-2021.appspot.com",
		sourceBucketPrefix:  os.Getenv("SOURCE_BUCKET_PREFIX"),  //for example: "data-mirror/",
		archiveBucketName:   os.Getenv("ARCHIVE_BUCKET_NAME"),   //for example: "data-dj-2021.appspot.com",
		archiveBucketPrefix: os.Getenv("ARCHIVE_BUCKET_PREFIX"), //for example: "data-archive/",
		sourceLocalDir:      os.Getenv("SOURCE_LOCAL_DIR"),      //for example: "../data/",
		archiveLocalDir:     os.Getenv("ARCHIVE_LOCAL_DIR"),     //for example: "../archives/",
		// port:                os.Getenv("PORT"),                  // Retrieve the PORT env variable for usage within the google cloud
		mode: os.Getenv("GIN_MODE"),
	}

	if cfg.port == "" {
		cfg.port = "8080"
		log.Printf("Defaulting to port %s", cfg.port)
	}

	if cfg.mode == "debug" {
		fmt.Printf("config: %#v\n", cfg)
	}

	return &cfg
}
