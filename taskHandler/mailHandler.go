package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	gomail "gopkg.in/mail.v2"
)

var (
	serviceEmailAddress  = os.Getenv("EMAIL_ADDRESS")
	serviceEmailPassword = os.Getenv("EMAIL_PASSWORD")
)

func handleEmailMessage(messagePayload string) {

	var archRequest archiveRequest

	// convert json string into struct
	json.Unmarshal([]byte(messagePayload), &archRequest)

	fmt.Println("handling email task: ", archRequest)
	err := sendEmail(archRequest)
	if err != nil {
		fmt.Println("err: ", err)
	}
}

func publishEmailTask(archRequest archiveRequest) error {

	//convert struct to json
	jsonMessage, err := json.Marshal(archRequest)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
		return err
	}
	//publish to channel
	channelName := "emails"
	err = rdbClient.Publish(channelName, jsonMessage).Err()
	if nil != err {
		fmt.Printf("Publish Error: %s", err.Error())
		return err
	}

	fmt.Println("published email task")
	return nil

}

//sendEmail sends an email with download link to user once downloading and zipping of the files is complete
func sendEmail(request archiveRequest) error {

	archFile := archBaseName + "_" + request.ArchiveID + ".zip" // name of the zip archive

	// construct content of the mail
	content := "The following files have been downloaded and were archived as " + archFile + ":\n\n"
	for _, name := range request.Files {
		content = content + name + "\n"
	}
	content = content + "\nThe archive can be retrieved from:\n" + "https://storage.googleapis.com/data-dj-2021.appspot.com/" + config.archiveBucketPrefix + archFile + "\n\nYours truly,\n\nThe DataDJ\n"

	// create new email message
	m := gomail.NewMessage()

	m.SetHeader("From", serviceEmailAddress)
	m.SetHeader("To", request.Email)
	m.SetHeader("Subject", "DataDJ Download completed")
	m.SetBody("text/plain", content)

	d := gomail.NewDialer("smtp.gmail.com", 587, serviceEmailAddress, serviceEmailPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("an error occured while sending the email notification")
	}
	fmt.Println("email sent successfully!")

	return nil
}