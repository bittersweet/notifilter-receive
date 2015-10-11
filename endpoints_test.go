package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
)

func init() {
	db, _ = sqlx.Connect("postgres", "user=markmulder dbname=notifier sslmode=disable")
}

func TestIndexStatus(t *testing.T) {
	url := "/"
	request, _ := http.NewRequest("GET", url, nil)
	response := httptest.NewRecorder()
	// need a resp.Body.Close() if it's a POST , in the handler, and a reader in the setup here:
	// http://play.golang.org/p/xpChdYyXWH
	// req, err := http.NewRequest("POST", "/", strings.NewReader("Hello"))

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
