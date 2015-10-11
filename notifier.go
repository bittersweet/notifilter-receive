package main

import (
	"bytes"
	"log"
	"text/template"

	"github.com/bittersweet/notifilter/notifiers"

	"github.com/jmoiron/sqlx/types"
)

// Notifier is a db-backed struct that contains everything that is necessary to
// check incoming events (rules) and what to do when those rules are matched.
type Notifier struct {
	ID               int            `db:"id"`
	EventName        string         `db:"event_name"`
	Template         string         `db:"template"`
	Rules            types.JsonText `db:"rules"`
	NotificationType string         `db:"notification_type"`
	Target           string         `db:"target"`
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

func (n *Notifier) getRules() []*rule {
	rules := []*rule{}
	n.Rules.Unmarshal(&rules)
	return rules
}

func (n *Notifier) checkRules(e *Event) bool {
	rules := n.getRules()
	rulesMet := true
	for _, rule := range rules {
		if !rule.Met(e) {
			log.Printf("Rule not met -- Key: %s, Type: %s, Setting %s, Value %s, Received Value %v\n", rule.Key, rule.Type, rule.Setting, rule.Value, e.toMap()[rule.Key])
			rulesMet = false
		}
	}
	if !rulesMet {
		log.Printf("Stopping notification of id: %d, rules not met\n", n.ID)
		return false
	}

	return true
}

func (n *Notifier) renderTemplate(e *Event) []byte {
	var err error
	var doc bytes.Buffer

	t := template.New("notificationTemplate")
	t, err = t.Parse(n.Template)
	if err != nil {
		log.Fatal("t.Parse of n.Template", err)
	}

	err = t.Execute(&doc, e.toMap())
	if err != nil {
		log.Fatal("t.Execute ", err)
	}

	return doc.Bytes()
}

func (n *Notifier) notify(e *Event, mn notifiers.MessageNotifier) {
	nt := n.NotificationType
	log.Printf("Notifying notifier id: %d type: %s\n", n.ID, nt)

	if !n.checkRules(e) {
		return
	}

	message := n.renderTemplate(e)
	mn.SendMessage(e.Identifier, n.Target, message)
	log.Printf("Notifying notifier id: %d done\n", n.ID)
}
