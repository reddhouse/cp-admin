package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func sendEmail() {
	emailSmtpServer := os.Getenv("EMAIL_SMTP_SERVER")
	emailPrimaryUser := os.Getenv("EMAIL_PRIMARY_USER")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	emailSenderAddress := os.Getenv("EMAIL_SENDER_ADDRESS")
	emailSenderName := os.Getenv("EMAIL_SENDER_NAME")
	emailTestRecipient := os.Getenv("EMAIL_TEST_RECIPIENT")

	auth := smtp.PlainAuth(
		"",
		emailPrimaryUser,
		emailPassword,
		emailSmtpServer,
	)

	to := []string{emailTestRecipient}
	msg := []byte("To: " + emailTestRecipient + "\r\n" +
		"Subject:  Cp-admin Update\r\n" +
		"From:  " + emailSenderName + " <" + emailSenderAddress + ">\r\n" +
		"Hello. Here is the update you requested!\r\n")

	err := smtp.SendMail(fmt.Sprintf("%s:587", emailSmtpServer), auth, emailSenderAddress, to, msg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Email sent to %v\n", emailTestRecipient)
}

func sendTLSEmail() {
	emailSmtpServer := os.Getenv("EMAIL_SMTP_SERVER")
	emailPrimaryUser := os.Getenv("EMAIL_PRIMARY_USER")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	emailSenderAddress := os.Getenv("EMAIL_SENDER_ADDRESS")
	emailSenderName := os.Getenv("EMAIL_SENDER_NAME")
	emailTestRecipient := os.Getenv("EMAIL_TEST_RECIPIENT")

	auth := smtp.PlainAuth(
		"",
		emailPrimaryUser,
		emailPassword,
		emailSmtpServer,
	)

	to := []string{emailTestRecipient}
	msg := []byte("To: " + emailTestRecipient + "\r\n" +
		"Subject: An encrypted message from cp-admin!\r\n" +
		"From:  " + emailSenderName + " <" + emailSenderAddress + ">\r\n" +
		"Hello. How's the weather?\r\n")

	// Establish plain text connection to SMTP server.
	c, err := smtp.Dial(fmt.Sprintf("%s:587", emailSmtpServer))
	if err != nil {
		log.Fatal(err)
	}
	// Upgrade to a secure connection using TLS.
	config := &tls.Config{ServerName: emailSmtpServer}
	if err = c.StartTLS(config); err != nil {
		log.Fatal(err)
	}
	// Authenticate.
	if err = c.Auth(auth); err != nil {
		log.Fatal(err)
	}
	// Specify the sender.
	if err = c.Mail(emailSenderAddress); err != nil {
		log.Fatal(err)
	}
	// Specify the recipients.
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			log.Fatal(err)
		}
	}
	// Get a writer from the server. Write message, and close.
	w, err := c.Data()
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(msg)
	if err != nil {
		log.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Email sent to %v\n", emailTestRecipient)
}
