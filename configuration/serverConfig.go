package configuration

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ServerConfig holds all the deployment specific environment variables and settings
// it gets initialised by a function and should be globally accessible
type ServerConfig struct {
	ProjectID            string // google cloud project-id that contains bucket resources
	ArchiveBaseURL       string // url where zip files will be accessible from
	SourceBucketName     string // name of bucket where provided files are stored
	SourceBucketPrefix   string // only returns files within this directory in the bucket
	ArchiveBucketName    string // name of bucket where archive files will be written to
	ArchiveBucketPrefix  string // path to directory in the bucket to write archive files to
	SourceLocalDir       string // path to directory on local machine with available files
	ArchiveLocalDir      string // path to directory on local machine to write archive files
	SourceAPIURL         string // URL to API giving access to provided files
	DbConnector          string // link to connect to mongoDB database
	DbName               string
	RdbHost              string // what address redis should connect to
	RdbPort              string // what port redis should connect to
	Port                 string // port to start api listening on
	Mode                 string // mode that gin is running in (use as mode for entire application)
	ServiceEmailAddress  string // email address for sending the form and download link
	ServiceEmailPassword string // password for email address
}

// InitServerConfig initializes the serverConfig struct with all the environment variables
func InitServerConfig() *ServerConfig {
	cfg := ServerConfig{
		ProjectID:            os.Getenv("PROJECT_ID"),            // for example: "data-dj-2021"
		ArchiveBaseURL:       os.Getenv("ARCHIVE_BASE_URL"),      // for example: "https://storage.googleapis.com/"
		SourceBucketName:     os.Getenv("SOURCE_BUCKET_NAME"),    // for example: "data-dj-2021.appspot.com",
		SourceBucketPrefix:   os.Getenv("SOURCE_BUCKET_PREFIX"),  // for example: "data-mirror/",
		ArchiveBucketName:    os.Getenv("ARCHIVE_BUCKET_NAME"),   // for example: "data-dj-2021.appspot.com",
		ArchiveBucketPrefix:  os.Getenv("ARCHIVE_BUCKET_PREFIX"), // for example: "data-archive/",
		SourceLocalDir:       os.Getenv("SOURCE_LOCAL_DIR"),      // for example: "../data/",
		ArchiveLocalDir:      os.Getenv("ARCHIVE_LOCAL_DIR"),     // for example: "../archives/",
		SourceAPIURL:         os.Getenv("SOURCE_API_URL"),        // for example: "",
		DbConnector:          os.Getenv("DB_CONNECTOR"),          // for example: "mongodb+srv://username:password@cluster.jzmhu.mongodb.net/collection?retryWrites=true&w=majority",
		DbName:               os.Getenv("DB_NAME"),               // for example: main or test
		RdbHost:              os.Getenv("REDISHOST"),             // for example: "0.0.0.0", "10.8.0.1" or "localhost",
		RdbPort:              os.Getenv("REDISPORT"),             // Usually 6379 for Redis
		Port:                 os.Getenv("PORT"),                  // retrieve the PORT env variable for usage within the google cloud,
		Mode:                 os.Getenv("GIN_MODE"),              // for example: "debug" or "production"
		ServiceEmailAddress:  os.Getenv("EMAIL_ADDRESS"),         // for example: "datadj.service@gmail.com"
		ServiceEmailPassword: os.Getenv("EMAIL_PASSWORD"),        // gotta find a good one yourself
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
		log.Printf("Defaulting to port %s", cfg.Port)
	}

	if cfg.Mode == "debug" {
		empJSON, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Print("config: \n", string(empJSON), "\n")

	}

	return &cfg
}
