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

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": "12"}`)
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

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": "12"}`)
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

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": "12"}`)
	s := Stat{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierNotify(t *testing.T) {
	// fake sendSlackNotification somehow?
}

func TestNotifierNotifyReturnsEarlyIfRulesAreNotMet(t *testing.T) {
}

func TestNotifierRenderTemplate(t *testing.T) {
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		Class:            "User",
		Template:         "name: {{.name}}",
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": "12"}`)
	s := Stat{"Mark", jt}

	result := n.renderTemplate(&s)
	expected := "name: Go"
	assert.Equal(t, result, expected)
}

func TestNotifierRenderTemplateWithLogic(t *testing.T) {
	template := `{{ if .active }}Active!{{ else }}inactive{{ end }}`
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		Class:            "User",
		Template:         template,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": "12"}`)
	s := Stat{"Mark", jt}

	result := n.renderTemplate(&s)
	expected := "Active!"
	assert.Equal(t, result, expected)

	jt = types.JsonText(`{"active": false, "name": "Go", "number": "12"}`)
	s = Stat{"Mark", jt}

	result = n.renderTemplate(&s)
	expected = "inactive"
	assert.Equal(t, result, expected)
}
