package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/jmoiron/sqlx/types"
)

type NotifierResponse struct {
	response *slackResponse
	error    error
}

type MessageNotifier interface {
	sendMessage(string, []byte) NotifierResponse
}

type Notifier struct {
	Id               int            `db:"id"`
	NotificationType string         `db:"notification_type"`
	Class            string         `db:"class"`
	Template         string         `db:"template"`
	Rules            types.JsonText `db:"rules"`
	// type slack/email/direct to phone
	// email address, slack channel, phone number, how to store?
}

func (n *Notifier) newNotifier() MessageNotifier {
	switch n.NotificationType {
	case "email":
		return &emailNotifier{}
	case "slack":
		return &slackNotifier{}
	}
	return &slackNotifier{}
}

func (n *Notifier) getRules() []*Rule {
	rules := []*Rule{}
	n.Rules.Unmarshal(&rules)
	return rules
}

func (n *Notifier) checkRules(s *Stat) bool {
	rules := n.getRules()
	rules_met := true
	for _, rule := range rules {
		if !rule.Met(s) {
			fmt.Printf("Rule not met -- Key: %s, Type: %s, Setting %s, Value %s, Received Value %v\n", rule.Key, rule.Type, rule.Setting, rule.Value, s.toMap()[rule.Key])
			rules_met = false
		}
	}
	if !rules_met {
		fmt.Printf("Stopping notification of id: %d, rules not met\n", n.Id)
		return false
	}

	return true
}

func (n *Notifier) renderTemplate(s *Stat) []byte {
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

func (n *Notifier) notify(s *Stat, mn MessageNotifier) {
	nt := n.NotificationType
	fmt.Printf("Notifying notifier id: %d type: %s\n", n.Id, nt)

	if !n.checkRules(s) {
		return
	}

	message := n.renderTemplate(s)
	mn.sendMessage(s.Key, message)
	fmt.Printf("Notifying notifier id: done\n", n.Id)
}
