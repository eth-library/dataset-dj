package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/eth-library/dataset-dj/mailHandler"

	"gopkg.in/yaml.v3"
)

func InitConfig(configPath string) *ApplicationConfig {
	var config ApplicationConfig
	yamlFile, err := os.ReadFile(configPath + "taskhandler.yml")
	if err != nil {
		log.Fatalln("critical", "No config found, will stop now.", err)
	}

	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalln("ERROR parsing config", fmt.Sprint(err))
	}

	return &config
}

func loadSecrets(configPath string) *Secrets {
	var secrets Secrets
	yamlFile, err := os.ReadFile(configPath + ".secrets.yml")
	if err != nil {
		log.Fatalln("critical", "No .secrets file at "+configPath+".secrets"+" found, will stop now.")
	}
	if err = yaml.Unmarshal(yamlFile, &secrets); err != nil {
		log.Fatalln("ERROR parsing secrets", fmt.Sprint(err))
	}
	return &secrets
}

func loadLibDriveConfig(configPath string) *LibDriveConfig {
	var config LibDriveConfig
	yamlFile, err := os.ReadFile(configPath + "libDrive.yml")
	if err != nil {
		log.Fatalln("critical", "No config found, will stop now.", err)
	}

	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalln("ERROR parsing config", fmt.Sprint(err))
	}

	return &config
}

func startDownloadLinkEmailTask(url string, recipientEmail string) {
	content := fmt.Sprintf(mailHandler.DownloadLinkContent, url, url)

	emailParts := mailHandler.EmailParts{
		To:         recipientEmail,
		Subject:    "DataDJ - Download complete - Link to retrieve requested files",
		BodyType:   "text/html",
		Body:       content,
		Server:     config.ServiceEmailHost,
		Address:    config.ServiceEmailAddress,
		Password:   config.ServiceEmailPassword,
		ErrorMsg:   "an error occurred while sending the download link email notification: ",
		SuccessMsg: "email with download link sent",
	}
	go mailHandler.SendEmailAsync(emailParts)
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func inTimeWindow() bool {
	now := getNow()
	return config.StartTime.IsZero() || config.EndTime.IsZero() || inTimeSpan(config.StartTime, config.EndTime, now)
}

func getNow() time.Time {
	nowBase := time.Now()
	now, err := time.Parse("15:04", fmt.Sprintf("%02d:%02d", nowBase.Hour(), nowBase.Minute()))
	if err != nil {
		println(err.Error())
		log.Fatal("Failed to assemble current time defined by layout (15:04)")
	}
	return now
}

func parseTimes(orders []dbutil.Order) []dbutil.TimedOrder {
	var res []dbutil.TimedOrder
	for _, o := range orders {
		date, _ := time.Parse(time.RFC822, o.Date)
		res = append(res, dbutil.TimedOrder{
			OrderID:   o.OrderID,
			ArchiveID: o.ArchiveID,
			Email:     o.Email,
			Date:      date,
			Status:    o.Status,
			Sources:   o.Sources,
		})
	}
	return res
}

// WriteLocalToZip is a helper function for writing an individual local file to zip.Writer object
func WriteLocalToZip(fileName string, writer *zip.Writer) error {

	if !fileExists(fileName) {
		return fmt.Errorf("file does not exist: %s\n", fileName)
	}
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("could not find file: %s\n%s", fileName, err)
	}

	w, err := writer.Create(fileName)
	if err != nil {
		return fmt.Errorf("could not create file in archive: %s", fileName)
	}

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("could not write to file in archive: %s", fileName)
	}
	return nil
}

func fileExists(fpath string) bool {
	_, err := os.Stat(fpath)
	if err == nil {
		return true
	}
	return false
}
