package notifiers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// SlackNotifier is a notifier accountable for sending notifications to Slack
type SlackNotifier struct {
	Token    string
	Username string
	Icon     string
}

// SendMessage sends an event with processed data to a selected Slack channel (target)
func (s *SlackNotifier) SendMessage(eventName string, target string, data []byte) {
	// https://api.slack.com/methods/chat.postMessage
	token := s.Token
	channel := target
	text := url.QueryEscape(string(data))
	username := s.Username
	icon := s.Icon

	format := "http://slack.com/api/chat.postMessage?token=%s&channel=%s&text=%s&username=%s&icon_url=%s"
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
	err = json.Unmarshal(contents, &slackResponse)
	if err != nil {
		log.Println("json.Unmarshal", err)
	}
	// TODO: Do something with slack response again, ok key or status code will
	// tell us if everything was alright
	fmt.Println("Slack response:", slackResponse)
}
