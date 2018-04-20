package notifiers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// SlackNotifier is a notifier accountable for sending notifications to Slack
type SlackNotifier struct {
	HookURL string
}

type SlackPayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// SendMessage sends an event with processed data to a selected Slack channel (target)
func (s *SlackNotifier) SendMessage(target string, eventName string, data []byte) {
	payload := SlackPayload{
		Channel: target,
		Text:    string(data),
	}

	payloadEnc, err := json.Marshal(payload)
	payloadReader := bytes.NewReader(payloadEnc)

	slackResp, err := http.Post(s.HookURL, "application/json", payloadReader)
	if err != nil {
		panic(err)
	}
	defer slackResp.Body.Close()

	slackBody, err := ioutil.ReadAll(slackResp.Body)
	log.Println("Slack Response:", string(slackBody))
}
