package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func Persist(key string, data map[string]interface{}) error {
	fmt.Printf("[ES] Persisting to %s: %v\n", key, data)

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
