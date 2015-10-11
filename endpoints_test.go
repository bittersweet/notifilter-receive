package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
)

func init() {
	db, _ = sqlx.Connect("postgres", "user=markmulder dbname=notifilter_development sslmode=disable")
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
