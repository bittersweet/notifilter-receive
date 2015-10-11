package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

func handleCount(w http.ResponseWriter, r *http.Request) {
	defer trackTime(time.Now(), "handleCount")

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM incoming").Scan(&count)
	if err != nil {
		log.Fatal("rowcount: ", err)
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
