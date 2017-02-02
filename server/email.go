package server

import (
	"log"
	"os"

	"errors"

	"github.com/cbroglie/mustache"
	"gopkg.in/mailgun/mailgun-go.v1"
)

// loadEmail loads both the HTML and raw text for a given email template and compiles them using
// mustache
func loadEmail(filename string, context map[string]string) (string, string, error) {
	txt, err := mustache.RenderFile("./server/emails/"+filename+".txt", context)
	if err != nil {
		return "", "", err
	}
	html, err := mustache.RenderFileInLayout("./server/emails/"+filename+".html.mustache",
		"./server/emails/layout.html.mustache", context)
	if err != nil {
		return "", "", err
	}
	return txt, html, nil
}

// Mailer is a wrapper around the Mailgun interface
type Mailer struct {
	mg   mailgun.Mailgun
	from string
}

// NewMailer initializes and returns a new MailGun object according the environment variables
func NewMailer(fromEmail string) (*Mailer, error) {
	domain := os.Getenv("MG_DOMAIN")
	if domain == "" {
		return nil, errors.New("Mailgun domain not defined")
	}
	apiKey := os.Getenv("MG_API_KEY")
	if apiKey == "" {
		return nil, errors.New("Mailgun API key not defined")
	}
	publicKey := os.Getenv("MG_PUBLIC_API_KEY")
	if publicKey == "" {
		return nil, errors.New("Mailgun public API key not defined")
	}
	mg := mailgun.NewMailgun(domain, apiKey, publicKey)
	m := Mailer{mg, fromEmail}
	return &m, nil
}

// MailSignup is meant to send the initial email when a new user signs up
func (M *Mailer) MailSignup(to, name string) (string, string, error) {
	context := map[string]string{"name": name}
	txt, html, err := loadEmail("signup", context)
	if err != nil {
		log.Panicln(err)
		return "", "", err
	}
	message := mailgun.NewMessage(
		M.from,
		"Welcome to AIComp",
		txt,
		to)
	message.SetHtml(html)
	resp, id, err := M.mg.Send(message)
	return resp, id, err
}
