package configuration

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultJobs     = 1
	defaultInterval = 900
	defaultURL      = "https://dj-api-ucooq6lz5a-oa.a.run.app"
	defaultLayout   = "15:04"
)

type HandlerConfig struct {
	MaxJobs         int
	RequestInterval int
	StartTime       time.Time
	EndTime         time.Time
	Sources         []string
	TargetURL       string
	HandlerKey      string
	Layout          string
}

func lookupTime(key string) time.Time {
	var timeVar time.Time
	timeString := os.Getenv(key)
	if timeString != "" {
		val, err := time.Parse(defaultLayout, timeString)
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

	// Handler Key
	handlerKey := lookupString("HANDLER_KEY", "")
	if handlerKey == "" {
		log.Fatal("HANDLER_KEY cannot be empty! Please request a handler from the API admin")
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
		MaxJobs:         jobs,
		RequestInterval: interval,
		StartTime:       startTime,
		EndTime:         endTime,
		Sources:         sources,
		TargetURL:       targetURL,
		HandlerKey:      handlerKey,
		Layout:          defaultLayout,
	}
	pretty, _ := json.MarshalIndent(hc, "", "  ")
	fmt.Print("config: \n", string(pretty), "\n")
	return &hc
}
