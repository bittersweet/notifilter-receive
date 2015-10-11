package main

import (
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

type LocalMessageNotifier struct {
	event_name string
	message    []byte
	processed  bool
}

func (mn *LocalMessageNotifier) sendMessage(event_name string, data []byte) NotifierResponse {
	mn.event_name = event_name
	mn.message = data
	mn.processed = true

	return NotifierResponse{}
}

func TestNewNotifier(t *testing.T) {
	n := Notifier{}
	assert.Equal(t, n.newNotifier(), &slackNotifier{})

	n.NotificationType = "email"
	assert.Equal(t, n.newNotifier(), &emailNotifier{})

	n.NotificationType = "slack"
	assert.Equal(t, n.newNotifier(), &slackNotifier{})
}

func TestNotifierCheckRulesSingle(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierCheckRulesMultiple(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"},
	{"key": "name", "type": "string", "setting": null, "value": "Go"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierCheckRulesSettingIsNull(t *testing.T) {
	var rules = types.JsonText(`[{"key": "name", "type": "string", "setting": null "value": "Go"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierCheckRulesSettingIsBlank(t *testing.T) {
	var rules = types.JsonText(`[{"key": "name", "type": "string", "setting": "", "value": "Go"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	assert.Equal(t, n.checkRules(&s), true)
}

func TestNotifierNotify(t *testing.T) {
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	mn := &LocalMessageNotifier{}
	n.notify(&s, mn)

	assert.Equal(t, mn.event_name, "Mark")
	assert.Equal(t, mn.message, []byte("name: Go"))
	assert.Equal(t, mn.processed, true)
}

func TestNotifierNotifyReturnsEarlyIfRulesAreNotMet(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "gt", "value": "1"}]`)
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 0}`)
	s := Event{"Mark", jt}

	mn := &LocalMessageNotifier{}
	n.notify(&s, mn)

	assert.Equal(t, mn.processed, false)
}

func TestNotifierRenderTemplate(t *testing.T) {
	n := Notifier{
		Id:               1,
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	result := n.renderTemplate(&s)
	expected := []byte("name: Go")
	assert.Equal(t, result, expected)
}

func TestNotifierRenderTemplateWithLogic(t *testing.T) {
	template := `{{ if .active }}Active!{{ else }}inactive{{ end }}`
	n := Notifier{
		Template: template,
	}

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	result := n.renderTemplate(&s)
	expected := []byte("Active!")
	assert.Equal(t, result, expected)

	jt = types.JsonText(`{"active": false, "name": "Go", "number": 12}`)
	s = Event{"Mark", jt}

	result = n.renderTemplate(&s)
	expected = []byte("inactive")
	assert.Equal(t, result, expected)
}

func TestNotifierRenderTemplateWithAdvancedLogic(t *testing.T) {
	n := Notifier{}
	n.Template = `Incoming conversion: {{ if gt .number 10.0 }}(Making it rain!) {{ end }}{{ .number }}`

	var jt = types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	s := Event{"Mark", jt}

	result := n.renderTemplate(&s)
	expected := "Incoming conversion: (Making it rain!) 12"
	assert.Equal(t, string(result), expected)

	jt = types.JsonText(`{"active": true, "name": "Go", "number": 10}`)
	s = Event{"Mark", jt}

	result = n.renderTemplate(&s)
	expected = "Incoming conversion: 10"
	assert.Equal(t, string(result), expected)
}
