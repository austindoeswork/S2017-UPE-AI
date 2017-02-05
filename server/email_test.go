package server

import (
	"log"
	"os"
	"testing"
)

func TestMailer(t *testing.T) {
	os.Chdir("..")
	Mailer, err := NewMailer("test@aicomp.io")
	if err != nil {
		log.Fatalln(err)
		t.FailNow()
	}
	var toEmail string
	if len(os.Args) > 2 {
		toEmail = os.Args[2]
	} else {
		toEmail = os.Args[1]
	}
	Mailer.MailSignup(toEmail, "Test User")
}
