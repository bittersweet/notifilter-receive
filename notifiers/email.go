package notifiers

import (
	"bytes"
	"log"
	"net/smtp"
	"text/template"
)

// EmailNotifier is a notifier accountable for e-mailing notifications
type EmailNotifier struct {
}

const emailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

<html>
<body>
{{.Body}}
</body>
</html>`

type emailData struct {
	From    string
	To      string
	Subject string
	Body    string
}

// SendMessage sends an event with processed data to a selected email address (target)
func (e *EmailNotifier) SendMessage(target string, eventName string, data []byte) {
	var err error
	var doc bytes.Buffer

	t := template.New("emailTemplate")
	t, err = t.Parse(emailTemplate)
	if err != nil {
		log.Fatal("t.Parse ", err)
	}
	context := &emailData{
		From:    "Springest Dev <developers@springest.nl>",
		To:      target,
		Subject: "Email subject line",
		Body:    string(data),
	}
	err = t.Execute(&doc, context)
	if err != nil {
		log.Fatal("t.Execute ", err)
	}

	// TODO: setup env variables to support multiple envs
	// TODO: Set up real test mode instead of using mailcatcher
	auth := smtp.PlainAuth("", "", "", "localhost:1025")
	err = smtp.SendMail("localhost:1025", auth, "test@example.com", []string{"recipient@example.com"}, doc.Bytes())
	if err != nil {
		log.Fatal("smtp.SendMail ", err)
	}
}
