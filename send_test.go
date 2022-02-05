package main

import (
	"fmt"
	"log"
	"net/smtp"
	"testing"
)

func TestSendMail(t *testing.T) {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial("127.0.0.1:2525")
	if err != nil {
		log.Fatal(err)
	}

	auth := smtp.PlainAuth("", "user@example.com", "password", "127.0.0.1")
	c.Auth(auth)

	// Set the sender and recipient first
	if err := c.Mail("sender@example.org"); err != nil {
		log.Fatal(err)
	}
	if err := c.Rcpt("recipient@example.net"); err != nil {
		log.Fatal(err)
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		log.Fatal(err)
	}
	_, err = fmt.Fprintf(wc, "To: recipient@example.net\r\n\r\nThis is the email body\r\n")
	if err != nil {
		log.Fatal(err)
	}
	err = wc.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		log.Fatal(err)
	}
}

func TestSendMailDirect(t *testing.T) {
	// Set up authentication information.
	// wrong host name: Leave out the port number (and colon) when passing the host name into
	// the auth mechanism. That should fix it.
	// https://groups.google.com/g/golang-nuts/c/5j1r43_Q4B8/m/eple2EjvWTYJ
	auth := smtp.PlainAuth("", "user@example.com", "password", "127.0.0.1")

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{"recipient@example.net"}
	msg := []byte("To: recipient@example.net\r\n" +
		"Subject: discount Gophers!\r\n" +
		"\r\n" +
		"This is the email body.\r\n")
	err := smtp.SendMail("127.0.0.1:2525", auth, "sender@example.org", to, msg)
	if err != nil {
		log.Fatal(err)
	}
}
