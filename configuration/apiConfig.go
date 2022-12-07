package configuration

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// ApiConfig holds the environment variable needed for deployment
type ApiConfig struct {
	ProjectID   string // google cloud project-id that contains bucket resources
	DbConnector string // link to connect to mongoDB database
	DbName      string // Name of Mongo database
	Port        string // port to start api listening on
	Mode        string // mode that gin is running in (use as mode for entire application)
}

func InitApiConfig() *ApiConfig {
	cfg := ApiConfig{
		ProjectID:   os.Getenv("PROJECT_ID"),   // for example: "data-dj-2021"
		DbConnector: os.Getenv("DB_CONNECTOR"), // for example: "mongodb+srv://username:password@cluster.jzmhu.mongodb.net/collection?retryWrites=true&w=majority",
		DbName:      os.Getenv("DB_NAME"),      // for example: main or test
		Port:        os.Getenv("PORT"),         // retrieve the PORT env variable for usage within the google cloud,
		Mode:        os.Getenv("GIN_MODE"),     // for example: "debug" or "production"
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
