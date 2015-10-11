package main

import (
	"fmt"
	"testing"

	"github.com/bittersweet/notifilter/notifiers"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

type LocalMessageNotifier struct {
	EventName string
	Target    string
	Message   []byte
	Processed bool
}

func (mn *LocalMessageNotifier) SendMessage(eventName string, target string, data []byte) {
	mn.EventName = eventName
	mn.Target = target
	mn.Message = data
	mn.Processed = true
}

func setupTestNotifier(data types.JsonText) Event {
	return Event{"signup", data}
}

func TestNewNotifier(t *testing.T) {
	n := Notifier{}
	assert.Equal(t, n.newNotifier(), &notifiers.SlackNotifier{})

	n.NotificationType = "email"
	assert.Equal(t, n.newNotifier(), &notifiers.EmailNotifier{})

	n.NotificationType = "slack"
	assert.Equal(t, n.newNotifier(), &notifiers.SlackNotifier{})
}

func TestNotifierCheckRulesSingle(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, n.checkRules(&event), true)
}

func TestNotifierCheckRulesMultiple(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"}, {"key": "name", "type": "string", "setting": null, "value": "Go"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, n.checkRules(&event), true)
}

func TestNotifierCheckRulesSettingIsNull(t *testing.T) {
	var rules = types.JsonText(`[{"key": "name", "type": "string", "setting": null "value": "Go"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, n.checkRules(&event), true)
}

func TestNotifierCheckRulesSettingIsBlank(t *testing.T) {
	var rules = types.JsonText(`[{"key": "name", "type": "string", "setting": "", "value": "Go"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, n.checkRules(&event), true)
}

func TestNotifierNotify(t *testing.T) {
	n := Notifier{
		EventName:        "User",
		Template:         "name: {{.name}}",
		NotificationType: "email",
		Target:           "email@example.com",
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	mn := &LocalMessageNotifier{}
	n.notify(&event, mn)

	assert.Equal(t, mn.EventName, "signup")
	assert.Equal(t, mn.Message, []byte("name: Go"))
	assert.Equal(t, mn.Processed, true)
}

func TestNotifierNotifyReturnsEarlyIfRulesAreNotMet(t *testing.T) {
	var rules = types.JsonText(`[{"key": "number", "type": "number", "setting": "gt", "value": "1"}]`)
	n := Notifier{
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
		NotificationType: "email",
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 0}`)
	event := setupTestNotifier(data)

	mn := &LocalMessageNotifier{}
	n.notify(&event, mn)

	assert.Equal(t, mn.Processed, false)
}

func TestNotifierRenderTemplate(t *testing.T) {
	n := Notifier{
		EventName:        "User",
		Template:         "name: {{.name}}",
		NotificationType: "email",
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, err := n.renderTemplate(&event)
	expected := []byte("name: Go")
	assert.Equal(t, result, expected)
	assert.Nil(t, err)
}

func TestNotifierRenderWithInvalidTemplate(t *testing.T) {
	n := Notifier{
		Template: "name: {{.name}",
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, err := n.renderTemplate(&event)
	expected := []byte("")
	assert.Equal(t, result, expected)
	assert.NotNil(t, err)
	assert.Equal(t, "template: notificationTemplate:1: unexpected \"}\" in operand", err.Error())
}

func TestNotifierRenderWithInvalidData(t *testing.T) {
	n := Notifier{
		Template: "name: {{.name}}",
	}

	data := types.JsonText(`{"active": true}`)
	event := setupTestNotifier(data)

	result, err := n.renderTemplate(&event)
	expected := []byte("name: <no value>")
	fmt.Println("resut, ", string(result))
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
}

func TestNotifierRenderTemplateWithLogic(t *testing.T) {
	template := `{{ if .active }}Active!{{ else }}inactive{{ end }}`
	n := Notifier{
		Template: template,
	}

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, _ := n.renderTemplate(&event)
	expected := []byte("Active!")
	assert.Equal(t, result, expected)

	data = types.JsonText(`{"active": false, "name": "Go", "number": 12}`)
	event = Event{"signup", data}

	result, _ = n.renderTemplate(&event)
	expected = []byte("inactive")
	assert.Equal(t, result, expected)
}

func TestNotifierRenderTemplateWithAdvancedLogic(t *testing.T) {
	n := Notifier{}
	n.Template = `Incoming conversion: {{ if gt .number 10.0 }}(Making it rain!) {{ end }}{{ .number }}`

	data := types.JsonText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, _ := n.renderTemplate(&event)
	expected := "Incoming conversion: (Making it rain!) 12"
	assert.Equal(t, string(result), expected)

	data = types.JsonText(`{"active": true, "name": "Go", "number": 10}`)
	event = setupTestNotifier(data)

	result, _ = n.renderTemplate(&event)
	expected = "Incoming conversion: 10"
	assert.Equal(t, string(result), expected)
}
