package main

import (
	"fmt"

	"github.com/jmoiron/sqlx/types"
)

type Notifier struct {
	Id               int            `db:"id"`
	NotificationType string         `db:"notification_type"`
	Class            string         `db:"class"`
	Template         string         `db:"template"`
	Rules            types.JsonText `db:"rules"`
	// type slack/email/direct to phone
	// email address, slack channel, phone number, how to store?
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
			fmt.Printf("Rule not met -- Key: %s, Type: %s, Setting %s, Value %s\n", rule.Key, rule.Type, rule.Setting, rule.Value)
			rules_met = false
		}
	}
	if !rules_met {
		fmt.Printf("Stopping notification of id: %d, rules not met\n", n.Id)
		return false
	}

	return true
}
