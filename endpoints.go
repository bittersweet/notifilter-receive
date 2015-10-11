package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handleCount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	count := countRows()
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
