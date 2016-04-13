package main

import (
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

func (mn *LocalMessageNotifier) SendMessage(target string, eventName string, data []byte) {
	mn.Target = target
	mn.EventName = eventName
	mn.Message = data
	mn.Processed = true
}

func setupTestNotifier(data types.JSONText) Event {
	return Event{
		Identifier: "signup",
		Data:       data,
	}
}

func TestNewNotifier(t *testing.T) {
	n := Notifier{}
	assert.Equal(t, &notifiers.SlackNotifier{}, n.newNotifier())

	n.NotificationType = "email"
	assert.Equal(t, &notifiers.EmailNotifier{}, n.newNotifier())

	n.NotificationType = "slack"
	assert.Equal(t, &notifiers.SlackNotifier{}, n.newNotifier())
}

func TestNotifierCheckRulesEmpty(t *testing.T) {
	var rules = types.JSONText(`[]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, true, n.checkRules(&event))
}

func TestNotifierCheckRulesSingle(t *testing.T) {
	var rules = types.JSONText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, true, n.checkRules(&event))
}

func TestNotifierCheckRulesMultiple(t *testing.T) {
	var rules = types.JSONText(`[{"key": "number", "type": "number", "setting": "eq", "value": "12"}, {"key": "name", "type": "string", "setting": null, "value": "Go"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, true, n.checkRules(&event))
}

func TestNotifierCheckRulesSettingIsNull(t *testing.T) {
	var rules = types.JSONText(`[{"key": "name", "type": "string", "setting": null "value": "Go"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, true, n.checkRules(&event))
}

func TestNotifierCheckRulesSettingIsBlank(t *testing.T) {
	var rules = types.JSONText(`[{"key": "name", "type": "string", "setting": "", "value": "Go"}]`)
	n := Notifier{
		NotificationType: "email",
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	assert.Equal(t, true, n.checkRules(&event))
}

func TestNotifierNotify(t *testing.T) {
	n := Notifier{
		EventName:        "signup",
		Template:         "name: {{.name}}",
		NotificationType: "email",
		Target:           "email@example.com",
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	mn := &LocalMessageNotifier{}
	n.notify(&event, mn)

	assert.Equal(t, "signup", mn.EventName)
	assert.Equal(t, []byte("name: Go"), mn.Message)
	assert.Equal(t, true, mn.Processed)
}

func TestNotifierNotifyReturnsEarlyIfRulesAreNotMet(t *testing.T) {
	var rules = types.JSONText(`[{"key": "number", "type": "number", "setting": "gt", "value": "1"}]`)
	n := Notifier{
		EventName:        "User",
		Template:         "name: {{.name}}",
		Rules:            rules,
		NotificationType: "email",
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 0}`)
	event := setupTestNotifier(data)

	mn := &LocalMessageNotifier{}
	n.notify(&event, mn)

	assert.Equal(t, false, mn.Processed)
}

func TestNotifierRenderTemplate(t *testing.T) {
	n := Notifier{
		EventName:        "User",
		Template:         "name: {{.name}}",
		NotificationType: "email",
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, err := n.renderTemplate(&event)
	expected := []byte("name: Go")
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
}

func TestNotifierRenderWithInvalidTemplate(t *testing.T) {
	n := Notifier{
		Template: "name: {{.name}",
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, err := n.renderTemplate(&event)
	expected := []byte("")
	assert.Equal(t, expected, result)
	assert.NotNil(t, err)
	assert.Equal(t, "template: notificationTemplate:1: unexpected \"}\" in operand", err.Error())
}

func TestNotifierRenderWithInvalidData(t *testing.T) {
	n := Notifier{
		Template: "name: {{.name}}",
	}

	data := types.JSONText(`{"active": true}`)
	event := setupTestNotifier(data)

	result, err := n.renderTemplate(&event)
	expected := []byte("name: <no value>")
	assert.Equal(t, expected, result)
	assert.Nil(t, err)
}

func TestNotifierRenderTemplateWithLogic(t *testing.T) {
	template := `{{ if .active }}Active!{{ else }}inactive{{ end }}`
	n := Notifier{
		Template: template,
	}

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, _ := n.renderTemplate(&event)
	expected := []byte("Active!")
	assert.Equal(t, expected, result)

	data = types.JSONText(`{"active": false, "name": "Go", "number": 12}`)
	event = setupTestNotifier(data)

	result, _ = n.renderTemplate(&event)
	expected = []byte("inactive")
	assert.Equal(t, expected, result)
}

func TestNotifierRenderTemplateWithAdvancedLogic(t *testing.T) {
	n := Notifier{}
	n.Template = `Incoming conversion: {{ if gt .number 10.0 }}(Making it rain!) {{ end }}{{ .number }}`

	data := types.JSONText(`{"active": true, "name": "Go", "number": 12}`)
	event := setupTestNotifier(data)

	result, _ := n.renderTemplate(&event)
	expected := "Incoming conversion: (Making it rain!) 12"
	assert.Equal(t, expected, string(result))

	data = types.JSONText(`{"active": true, "name": "Go", "number": 10}`)
	event = setupTestNotifier(data)

	result, _ = n.renderTemplate(&event)
	expected = "Incoming conversion: 10"
	assert.Equal(t, expected, string(result))
}

func TestNotifierRenderTemplateWithIsset(t *testing.T) {
	n := Notifier{}
	n.Template = `{{ if isset . "number" }}number ({{ .number }}) is set!{{ end }}{{ if isset . "nope" }} this will not show {{ end }}`

	data := types.JSONText(`{"number": 12}`)
	event := setupTestNotifier(data)

	result, err := n.renderTemplate(&event)
	if err != nil {
		t.Fatal(err)
	}
	expected := "number (12) is set!"
	assert.Equal(t, expected, string(result))
}

func TestIsset(t *testing.T) {
	data := map[string]interface{}{"a": 1}

	assert.True(t, isset(data, "a"))
	assert.False(t, isset(data, "b"))
}

func TestPresent(t *testing.T) {
	assert.False(t, present(nil))
	assert.False(t, present(""))
	assert.False(t, present(false))
	assert.True(t, present("not blank"))
	assert.True(t, present(true))
}

func TestEq(t *testing.T) {
	data := map[string]interface{}{"a": "string"}

	assert.True(t, eq(data["a"], "string"))
	assert.False(t, eq(data["a"], "something else"))
}
