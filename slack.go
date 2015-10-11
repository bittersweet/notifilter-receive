package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type slackNotifier struct {
}

type slackResponse struct {
	Ok bool   `json:"ok"`
	Ts string `json:"ts"`
}

var transport http.RoundTripper

func getTransport() http.RoundTripper {
	// If we have overridden this variable in testing
	if transport != nil {
		return transport
	}
	return http.DefaultTransport
}

func SetTransport(t http.RoundTripper) {
	transport = t
}

// TODO: Get settings from env variables
func (s *slackNotifier) sendMessage(event_name string, data []byte) NotifierResponse {
	client := &http.Client{Transport: getTransport()}
	// https://api.slack.com/methods/chat.postMessage
	// format := "http://slack.com/api/chat.postMessage?token=%s&channel=%s&text=%s&username=%s&icon_url=%s"
	format := "http://slack.com/api/chat.postMessage?token=%s&channel=%s&text=%s&username=%s&icon_url=%s"
	// format := "http://localhost:8000/?token=%s&channel=%s&text=%s&username=%s&icon_url=%s"
	// TODO: get an application token instead of using a personal one
	token := "xoxp-2152199637-2401973798-4106776238-a6cdd3"
	channel := "C0434MV7L" // mark-test
	text := url.QueryEscape(string(data))
	username := "Notifier"
	// TODO: Set a custom avatar
	icon := "http://lorempixel.com/48/48/"
	url := fmt.Sprintf(format, token, channel, text, username, icon)
	res, err := client.Get(url)
	if err != nil {
		log.Println("http.Get", err)
		return NotifierResponse{&slackResponse{}, err}
	}
	defer res.Body.Close()
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("ReadAll", err)
	}

	slackResponse := &slackResponse{}
	err = json.Unmarshal(contents, slackResponse)
	if err != nil {
		log.Println("json.Unmarshal", err)
	}

	return NotifierResponse{slackResponse, err}
}
