package main

import (
	"bytes"
	"log"
	"text/template"

	"github.com/bittersweet/notifilter/notifiers"

	"github.com/jmoiron/sqlx/types"
)

type Notifier struct {
	Id               int            `db:"id"`
	NotificationType string         `db:"notification_type"`
	EventName        string         `db:"event_name"`
	Template         string         `db:"template"`
	Rules            types.JsonText `db:"rules"`
	// type slack/email/direct to phone
	// email address, slack channel, phone number, how to store?
}

func (n *Notifier) newNotifier() notifiers.MessageNotifier {
	switch n.NotificationType {
	case "email":
		return &notifiers.EmailNotifier{}
	case "slack":
		return &notifiers.SlackNotifier{}
	}
	return &notifiers.SlackNotifier{}
}

func (n *Notifier) getRules() []*Rule {
	rules := []*Rule{}
	n.Rules.Unmarshal(&rules)
	return rules
}

func (n *Notifier) checkRules(e *Event) bool {
	rules := n.getRules()
	rules_met := true
	for _, rule := range rules {
		if !rule.Met(e) {
			log.Printf("Rule not met -- Key: %s, Type: %s, Setting %s, Value %s, Received Value %v\n", rule.Key, rule.Type, rule.Setting, rule.Value, e.toMap()[rule.Key])
			rules_met = false
		}
	}
	if !rules_met {
		log.Printf("Stopping notification of id: %d, rules not met\n", n.Id)
		return false
	}

	return true
}

func (n *Notifier) renderTemplate(s *Event) []byte {
	var err error
	var doc bytes.Buffer

	t := template.New("notificationTemplate")
	t, err = t.Parse(n.Template)
	if err != nil {
		log.Fatal("t.Parse of n.Template", err)
	}

	err = t.Execute(&doc, s.toMap())
	if err != nil {
		log.Fatal("t.Execute ", err)
	}

	return doc.Bytes()
}

func (n *Notifier) notify(s *Event, mn notifiers.MessageNotifier) {
	nt := n.NotificationType
	log.Printf("Notifying notifier id: %d type: %s\n", n.Id, nt)

	if !n.checkRules(s) {
		return
	}

	message := n.renderTemplate(s)
	mn.SendMessage(s.Identifier, message)
	log.Printf("Notifying notifier id: %d done\n", n.Id)
}
