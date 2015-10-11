// Package elasticsearch provides basic commands to interact with Elasticsearch
package elasticsearch

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ESPayload struct {
	Key        string                 `json:"key"`
	ReceivedAt time.Time              `json:"received_at"`
	Data       map[string]interface{} `json:"data"`
}

// Persist saves incoming events to Elasticsearch
func Persist(key string, data map[string]interface{}) error {
	log.Printf("[ES] Persisting to %s: %v\n", key, data)

	payload := ESPayload{
		Key:        key,
		ReceivedAt: time.Now(),
		Data:       data,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:9200/notifilter/event/?pretty", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("[ES] Success")
	log.Println("[ES]", string(body))
	return nil
}

// EventCount returns the total amount of events persisted to Elasticsearch
func EventCount() (int, error) {
	type response struct {
		Hits struct {
			Total int `json:"total"`
		} `json:"hits"`
	}

	resp, err := http.Get("http://localhost:9200/notifilter/event/_search?size=0")
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	log.Println("[ES] response: ", string(body))

	var parsed response
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return 0, err
	}

	return parsed.Hits.Total, nil
}
