// Package elasticsearch provides basic commands to interact with Elasticsearch
package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Client struct {
	Host  string
	Port  int
	Index string
}

type ElasticsearchClient interface {
	EventCount() (int, error)
}

// URL returns the full URL to the elasticsearch host, port and index
func (c *Client) URL() string {
	return fmt.Sprintf("http://%s:%d/%s", c.Host, c.Port, c.Index)
}

// Persist saves incoming events to Elasticsearch
func (c *Client) Persist(requestID string, application string, name string, data map[string]interface{}) error {
	log.Printf("%s [ES] persisting application=%s event=%s data=%v\n", requestID, application, name, data)

	payload := struct {
		Application string                 `json:"application"`
		Name        string                 `json:"name"`
		ReceivedAt  time.Time              `json:"received_at"`
		Data        map[string]interface{} `json:"data"`
	}{
		Application: application,
		Name:        name,
		ReceivedAt:  time.Now(),
		Data:        data,
	}

	reqBody, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/_doc", c.URL())
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		log.Printf("[ES] Error on persisting: %s", string(body))
		return errors.New("Failure on persist")
	}

	log.Printf("%s [ES] persisted\n", requestID)
	return nil
}

// EventCount returns the total amount of events persisted to Elasticsearch
func (c *Client) EventCount() (int, error) {
	type response struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}

	resp, err := http.Get(c.URL() + "/_search?size=0")
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

	return parsed.Hits.Total.Value, nil
}
