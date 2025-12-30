package models

import (
	"fmt"
	"log"

	"github.com/wneessen/go-mail"
)

const (
	DefaultSender = "test@lenslocked.com"
)

type Email struct {
	From      string
	To        string
	Subject   string
	PlainText string
	HTML      string
}

type EmailService struct {
	DefaultSender string
	client        *mail.Client
}

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
}

func NewEmailService(config SMTPConfig) (*EmailService, error) {
	client, err := mail.NewClient(
		config.Host,
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(config.User),
		mail.WithPassword(config.Pass),
		mail.WithPort(config.Port),
	)
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
		return nil, err
	}
	es := EmailService{
		DefaultSender: DefaultSender,
		client:        client,
	}

	return &es, nil
}

func (es *EmailService) Send(email Email) error {
	msg := mail.NewMsg()
	// From
	es.setFrom(msg, email)
	// To
	if err := msg.To(email.To); err != nil {
		log.Fatalf("failed to set To address: %s", err)
		return fmt.Errorf("failed to set To address: %s", err)
	}
	// Subject
	msg.SetGenHeader("Subject", email.Subject)
	// Body
	switch {
	case email.PlainText != "" && email.HTML != "":
		msg.SetBodyString(mail.TypeTextPlain, email.PlainText)
		msg.AddAlternativeString(mail.TypeTextHTML, email.HTML)
	case email.PlainText != "":
		msg.SetBodyString(mail.TypeTextPlain, email.PlainText)
	case email.HTML != "":
		msg.AddAlternativeString(mail.TypeTextHTML, email.HTML)
	}
	// Send it
	if err := es.client.DialAndSend(msg); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (es *EmailService) setFrom(msg *mail.Msg, email Email) {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}
	msg.From(from)
}
