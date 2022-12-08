package mailHandler

import (
	"crypto/tls"
	"fmt"
	gomail "gopkg.in/mail.v2"
)

// EmailParts of a prepared email, queued for sending
type EmailParts struct {
	To         string
	Subject    string
	BodyType   string // e.g.: text/plain
	Body       string
	Server     string
	Address    string
	Password   string
	ErrorMsg   string
	SuccessMsg string
}

// SendEmailAsync creates a new message object from the EmailParts and sends it asynchronously
func SendEmailAsync(email EmailParts) {
	// create new email message
	m := gomail.NewMessage()

	m.SetHeader("From", email.Address)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)
	m.SetBody(email.BodyType, email.Body)

	d := gomail.NewDialer(email.Server, 587, email.Address, email.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(fmt.Errorf(email.ErrorMsg + err.Error()))
	}
	fmt.Println(email.SuccessMsg)
}
