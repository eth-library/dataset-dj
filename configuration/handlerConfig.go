package configuration

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultJobs       = 1
	defaultInterval   = 900
	defaultURL        = "https://dj-api-ucooq6lz5a-oa.a.run.app"
	defaultArchiveDir = "./results"
)

type HandlerConfig struct {
	MaxJobs              int
	RequestInterval      int
	StartTime            time.Time
	EndTime              time.Time
	Sources              []string
	TargetURL            string
	HandlerKey           string
	ServiceEmailHost     string // The email server to connect to
	ServiceEmailAddress  string // email address for sending the form and download link
	ServiceEmailPassword string // password for email address
	ArchiveDir           string
}

func lookupTime(key string) time.Time {
	var timeVar time.Time
	timeString := os.Getenv(key)
	if timeString != "" {
		val, err := time.Parse("15:04", timeString)
		if err != nil {
			log.Fatal(key + " could not be converted to correct time format")
		}
		timeVar = val
	}
	return timeVar
}

func lookupNumerical(key string, def int) int {
	tmp := def
	tmpString := os.Getenv(key)
	if tmpString != "" {
		val, err := strconv.Atoi(tmpString)
		if err != nil || val == 0 {
			log.Fatal(key + " has to be an integer larger than 0")
		}
		tmp = val
	}
	return tmp
}

func lookupString(key string, def string) string {
	tmp := def
	tmpString := os.Getenv(key)
	if tmpString != "" {
		tmp = tmpString
	}
	return tmp
}

func InitHandlerConfig() *HandlerConfig {
	// Amount of  jobs
	jobs := lookupNumerical("MAX_JOBS", defaultJobs)

	// Time between requests to API endpoint
	interval := lookupNumerical("REQUEST_INTERVAL", defaultInterval)

	// Start time
	startTime := lookupTime("START_TIME")

	// End time
	endTime := lookupTime("END_TIME")

	// API URL
	targetURL := lookupString("TARGET_URL", defaultURL)

	// Handler key
	handlerKey := lookupString("HANDLER_KEY", "")
	if handlerKey == "" {
		log.Fatal("HANDLER_KEY cannot be empty! Please request a handler from the API admin")
	}

	// Email host
	emailHost := lookupString("EMAIL_HOST", "")
	if emailHost == "" {
		log.Fatal("EMAIL_HOST cannot be empty!")
	}

	// Email address
	emailAddress := lookupString("EMAIL_ADDRESS", "")
	if emailAddress == "" {
		log.Fatal("EMAIL_ADDRESS cannot be empty!")
	}

	// Email password
	emailPassword := lookupString("EMAIL_PASSWORD", "")
	if emailPassword == "" {
		log.Fatal("EMAIL_PASSWORD cannot be empty!")
	}

	// Directory where zipped archives are stored
	archiveDir := lookupString("ARCHIVE_DIR", defaultArchiveDir)
	if archiveDir == "" {
		log.Fatal("ARCHIVE_DIR cannot be empty!")
	}

	// Source Ids
	var sources []string
	srcAmount, err := strconv.Atoi(os.Getenv("SOURCE_AMOUNT"))
	if err != nil || srcAmount == 0 {
		log.Fatal("SOURCE_AMOUNT has to be an integer larger than 0")
	}
	for i := 0; i < srcAmount; i++ {
		sources = append(sources, os.Getenv("SOURCE_"+strings.ToUpper(strconv.Itoa(i))))
	}

	// Create config struct
	hc := HandlerConfig{
		MaxJobs:              jobs,
		RequestInterval:      interval,
		StartTime:            startTime,
		EndTime:              endTime,
		Sources:              sources,
		TargetURL:            targetURL,
		HandlerKey:           handlerKey,
		ServiceEmailHost:     emailHost,     // for example: "smtp.gmail.com"
		ServiceEmailAddress:  emailAddress,  // for example: "datadj.service@gmail.com"
		ServiceEmailPassword: emailPassword, // gotta find a good one yourself
		ArchiveDir:           archiveDir,
	}
	// pretty, _ := json.MarshalIndent(hc, "", "  ")
	// fmt.Print("config: \n", string(pretty), "\n")
	return &hc
}
