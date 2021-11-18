package main

import (
	"crypto/tls"
	"fmt"

	gomail "gopkg.in/mail.v2"
)

func sendNotification(request archiveRequest) error {

	content := "The following files have been downloaded and were archived as " + archBaseName + "_" + request.ArchiveID + ".zip:\n\n"
	for _, name := range request.Files {
		content = content + name + "\n"
	}
	content = content + "\nThe archive can be retrieved from " + storage + "\n\nYours truly,\n\nThe DataDJ\n"

	m := gomail.NewMessage()

	m.SetHeader("From", "datadj.service@gmail.com")
	m.SetHeader("To", request.Email)
	m.SetHeader("Subject", "DataDJ Download completed")
	m.SetBody("text/plain", content)

	d := gomail.NewDialer("smtp.gmail.com", 587, "datadj.service@gmail.com", password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("an error occured while sending the notification")
	}
	fmt.Println("Email sent successfully!")

	return nil
}
