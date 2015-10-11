package main

import (
	"fmt"
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

func TestCountStatus(t *testing.T) {
	mockES := MockElasticsearchClient{}
	countHandle := handleCount(mockES)
	request, _ := http.NewRequest("GET", "/v1/count", nil)
	response := httptest.NewRecorder()
	countHandle.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	fmt.Println(response.Body)
}
