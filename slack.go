package main

import (
	"bytes"
	"fmt"
	// "io/ioutil"
	"log"
	"net/http"
	"net/url"
	"text/template"
)

func sendSlackNotification(s *Stat, notifier *dbNotifier) {
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

// TODO: Get settings from env variables
func sendSlack(class string, data []byte) {
	// https://api.slack.com/methods/chat.postMessage
	format := "https://slack.com/api/chat.postMessage?token=%s&channel=%s&text=%s&username=%s&icon_url=%s"
	// TODO: get an application token instead of using a personal one
	token := "xoxp-2152199637-2401973798-4106776238-a6cdd3"
	channel := "C0434MV7L" // mark-test
	text := url.QueryEscape(string(data))
	username := "Notifier"
	// TODO: Set a custom avatar
	icon := "http://lorempixel.com/48/48/"
	url := fmt.Sprintf(format, token, channel, text, username, icon)
	res, err := http.Get(url)
	if err != nil {
		log.Println("http.Get", err)
	}
	defer res.Body.Close()
	// contents, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	log.Println("ReadAll", err)
	// }
	// fmt.Println(string(contents))
}
