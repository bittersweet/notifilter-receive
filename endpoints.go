package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bittersweet/notifilter/elasticsearch"
)

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

func handleCount(w http.ResponseWriter, r *http.Request) {
	defer trackTime(time.Now(), "handleCount")

	count, err := elasticsearch.EventCount()
	if err != nil {
		log.Fatal("Error while getting count from ES", err)
	}

	jmap := map[string]int{
		"status": 200,
		"count":  count,
	}

	output, err := json.MarshalIndent(jmap, "", "  ")
	if err != nil {
		log.Fatal("MarshalIndent", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(output)
}
