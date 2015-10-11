package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"text/template"
)

var transport http.RoundTripper

type slackResponse struct {
	Ok bool   `json:"ok"`
	Ts string `json:"ts"`
}

func sendSlackNotification(s *Stat, notifier *Notifier) {
	var err error
	var doc bytes.Buffer

	t := template.New("notificationTemplate")
	t, err = t.Parse(notifier.Template)
	if err != nil {
		log.Fatal("t.Parse of n.Template", err)
	}

	err = t.Execute(&doc, s.toMap())
	if err != nil {
		log.Fatal("t.Execute ", err)
	}

	sendSlack(s.Key, doc.Bytes())
}

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
func sendSlack(class string, data []byte) (*slackResponse, error) {
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
		return &slackResponse{}, err
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

	return slackResponse, err
}
