package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func closeTestServer(test_server *httptest.Server) {
	transport = nil
	test_server.Close()
}

func createTestServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	test_server := httptest.NewServer(http.HandlerFunc(handler))

	SetTransport(&http.Transport{Proxy: func(*http.Request) (*url.URL, error) { return url.Parse(test_server.URL) }})

	return test_server
}

func returnTestResponseForPath(path string, dummy_response string, t *testing.T) *httptest.Server {
	dummy_data := []byte(dummy_response)

	return createTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Error("Path doesn't match")
		}

		w.Write(dummy_data)
	})
}

func TestSendSlack(t *testing.T) {
	var resp string = `
{
	"ok": true,
	"channel": "C0434MV7L",
	"ts": "1405895017.000506",
	"message": {
		"text": "User mark created",
		"username": "Notifier",
		"type": "message",
		"subtype": "bot_message"
	}
}
`
	test_server := returnTestResponseForPath("/api/chat.postMessage", resp, t)
	defer closeTestServer(test_server)

	response, _ := sendSlack("mark", []byte{10, 10})
	if response.Ok != true {
		t.Error("response is not correct")
	}
}
