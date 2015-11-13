package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockElasticsearchClient struct {
}

func (mec MockElasticsearchClient) EventCount() (int, error) {
	return 1, nil
}

func (mec MockElasticsearchClient) Persist(a string, b string, c string, d map[string]interface{}) error {
	return nil
}

func (mec MockElasticsearchClient) GetEvent(a string, b string) ([]byte, error) {
	return []byte(""), nil
}

func TestCountStatus(t *testing.T) {
	mockES := MockElasticsearchClient{}
	countHandle := handleCount(mockES)
	request, _ := http.NewRequest("GET", "/v1/count", nil)
	response := httptest.NewRecorder()
	countHandle.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
}
