package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func init() {
	db, _ = sqlx.Connect("postgres", "user=markmulder dbname=notifier sslmode=disable")
}

func TestIndexStatus(t *testing.T) {
	url := "/"
	request, _ := http.NewRequest("GET", url, nil)
	response := httptest.NewRecorder()

	handleIndex(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Response body did not contain expected %v:\n\tbody: %v", "200", response.Code)
	}
}

func TestCountStatus(t *testing.T) {
	url := "/v1/count"
	request, _ := http.NewRequest("GET", url, nil)
	response := httptest.NewRecorder()
	handleCount(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Response body did not contain expected %v:\n\tbody: %v", "200", response.Code)
	}
}

func TestCreateWithInvalidBody(t *testing.T) {
	url := "/create"
	request, _ := http.NewRequest("POST", url, strings.NewReader(""))
	response := httptest.NewRecorder()
	handleCreate(response, request)

	assert.Equal(t, 500, response.Code)
}

func TestCreateNotifier(t *testing.T) {
	params := url.Values{"notification_type": {"slack"}, "class": {"User"}, "template": {"test"}, "rules": {"{}"}}
	request, _ := http.NewRequest("POST", "/create", strings.NewReader(params.Encode()))
	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded; param=value",
	)
	response := httptest.NewRecorder()
	handleCreate(response, request)

	assert.Equal(t, 302, response.Code)

	var notifier Notifier
	err := db.Get(&notifier, "SELECT * FROM notifiers ORDER BY id DESC LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, "slack", notifier.NotificationType)
	assert.Equal(t, "User", notifier.Class)
	assert.Equal(t, "test", notifier.Template)
	assert.Equal(t, "{}", notifier.Rules)
}

func TestPreview(t *testing.T) {
	var incomingId int
	query := `INSERT INTO incoming(received_at, class, data) VALUES($1, $2, $3) RETURNING id`
	data := `{"name": "Mark", "number": 100}`
	err := db.QueryRow(query, time.Now(), "Booking", data).Scan(&incomingId)
	if err != nil {
		t.Fatal("Insert in TestPreview", err)
	}

	params := url.Values{"class": {"Booking"}, "template": {"{{.name}} is pretty cool, the number of the day is: {{.number}}"}}
	request, _ := http.NewRequest("POST", "/preview", strings.NewReader(params.Encode()))
	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded; param=value",
	)
	response := httptest.NewRecorder()
	handlePreview(response, request)

	assert.Equal(t, 200, response.Code)

	result := string(response.Body.Bytes())
	expected := "Mark is pretty cool, the number of the day is: 100"
	assert.Equal(t, string(result), expected)
}

func TestPreviewWithInvalidInput(t *testing.T) {
	params := url.Values{"class": {"Booking"}, "template": {"{{.name"}}
	request, _ := http.NewRequest("POST", "/preview", strings.NewReader(params.Encode()))
	request.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded; param=value",
	)
	response := httptest.NewRecorder()
	handlePreview(response, request)

	assert.Equal(t, 204, response.Code)
}
