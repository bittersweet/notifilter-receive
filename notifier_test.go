package main

import (
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

func TestNotifierCheckRulesSingle(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		Class:            "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Stat{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierCheckRulesMultiple(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"},
	{"key": "name", "type": "string", "setting": null, "value": "Go"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		Class:            "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Stat{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierCheckRulesSettingIsNull(t *testing.T) {
	var rules = types.JsonText(`[{"key": "name", "type": "string", "setting": null "value": "Go"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		Class:            "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Stat{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierCheckRulesSettingIsBlank(t *testing.T) {
	var rules = types.JsonText(`[{"key": "name", "type": "string", "setting": "", "value": "Go"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		Class:            "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Stat{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}
