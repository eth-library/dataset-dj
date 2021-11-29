package main

import (
	"log"
	"os"
)

type serverConfig struct {
	projectID           string // google cloud project-id that contains bucket resources
	sourceBucketName    string // name of bucket where provided files are stored
	sourceBucketPrefix  string // only returns files within this directory in the bucket
	archiveBucketName   string // name of bucket where archive files will be written to
	archiveBucketPrefix string // path to directory in the bucket to write archive files to
	sourceLocalDir      string // path to directory on local machine with available files
	archiveLocalDir     string // path to directory on local machine to write archive files
	port                string // port to start api listening on
}

func initServerConfig() *serverConfig {
	sc := serverConfig{
		projectID:           "data-dj-2021",
		sourceBucketName:    "data-dj-2021.appspot.com",
		sourceBucketPrefix:  "data-mirror/",
		archiveBucketName:   "data-dj-2021.appspot.com",
		archiveBucketPrefix: "data-archive/",
		sourceLocalDir:      "../data/",
		archiveLocalDir:     "../archives/",
		port:                os.Getenv("PORT"), // Retrieve the PORT env variable for usage within the google cloud
	}

	if sc.port == "" {
		sc.port = "8080"
		log.Printf("Defaulting to port %s", sc.port)
	}

	return &sc
}
