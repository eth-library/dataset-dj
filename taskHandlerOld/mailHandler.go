package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	gomail "gopkg.in/mail.v2"
)

//EmailParts of a prepared email, queued for sending
type EmailParts struct {
	To       string
	Subject  string
	BodyType string // e.g.: text/plain
	Body     string
}

func handleEmailMessage(messagePayload string) {

	var emailParts EmailParts

	// convert json string into struct
	json.Unmarshal([]byte(messagePayload), &emailParts)

	err := sendEmail(emailParts)
	if err != nil {
		fmt.Println("err: ", err)
	}
}

//sendEmail creates a new message object from the EmailParts and sends it
func sendEmail(email EmailParts) error {
	// create new email message
	m := gomail.NewMessage()

	m.SetHeader("From", config.ServiceEmailAddress)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)
	m.SetBody(email.BodyType, email.Body)

	d := gomail.NewDialer("smtp.gmail.com", 587, config.ServiceEmailAddress, config.ServiceEmailPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("an error occured while sending the email notification: " + err.Error())
	}
	fmt.Println("email sent successfully!")

	return nil
}
