package notifiers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type SlackNotifier struct {
}

func (s *SlackNotifier) SendMessage(event_name string, data []byte) {
	// https://api.slack.com/methods/chat.postMessage
	format := "http://slack.com/api/chat.postMessage?token=%s&channel=%s&text=%s&username=%s&icon_url=%s"
	// TODO: get an application token instead of using a personal one
	token := "xoxp-2152199637-2401973798-4106776238-a6cdd3"
	channel := "C0434MV7L" // mark-test
	text := url.QueryEscape(string(data))
	username := "Notifier"
	icon := "http://lorempixel.com/48/48/"
	url := fmt.Sprintf(format, token, channel, text, username, icon)

	res, err := http.Get(url)
	if err != nil {
		log.Println("http.Get", err)
	}
	defer res.Body.Close()

	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("ReadAll", err)
	}

	slackResponse := struct {
		Ok bool   `json:"ok"`
		Ts string `json:"ts"`
	}{}
	err = json.Unmarshal(contents, slackResponse)
	if err != nil {
		log.Println("json.Unmarshal", err)
	}
}
