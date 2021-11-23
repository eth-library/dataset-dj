package main

import (
	"crypto/tls"
	"fmt"

	gomail "gopkg.in/mail.v2"
)

// send email with download link to user once downloading and zipping of the files is complete
func sendNotification(request archiveRequest) error {

	archFile := archBaseName + "_" + request.ArchiveID + ".zip" // name of the zip archive

	// construct content of the mail
	content := "The following files have been downloaded and were archived as " + archFile + ":\n\n"
	for _, name := range request.Files {
		content = content + name + "\n"
	}
	content = content + "\nThe archive can be retrieved from:\n" + "https://storage.googleapis.com/data-dj-2021.appspot.com/" + archStorage + archFile + "\n\nYours truly,\n\nThe DataDJ\n"

	// create new email message
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
	fmt.Println("email sent successfully!")

	return nil
}
