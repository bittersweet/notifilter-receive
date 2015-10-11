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

func handleCount(es elasticsearch.ElasticsearchClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer trackTime(time.Now(), "handleCount")

		count, err := es.EventCount()
		if err != nil {
			log.Println("Error while getting count from ES", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jmap := map[string]int{
			"status": 200,
			"count":  count,
		}

		output, err := json.MarshalIndent(jmap, "", "  ")
		if err != nil {
			log.Println("Error in /v1/count MarshalIndent", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(output)
	})
}
