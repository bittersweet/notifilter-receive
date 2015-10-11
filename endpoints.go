package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"
)

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	defer trackTime(time.Now(), "handleIndex")
	defer r.Body.Close()

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("ParseFiles", err)
	}

	notifiers := []Notifier{}
	err = db.Select(&notifiers, "SELECT * FROM notifiers")
	if err != nil {
		log.Fatal("db.Select notifiers ", err)
	}

	incoming := []Incoming{}
	err = db.Select(&incoming, "SELECT * FROM incoming ORDER BY id desc")
	if err != nil {
		log.Fatal("db.Select incoming ", err)
	}

	data := map[string]interface{}{
		"notifiers": notifiers,
		"incoming":  incoming,
	}
	err = t.Execute(w, data)
	if err != nil {
		log.Fatal("t.Execute", err)
	}
}

func handleCount(w http.ResponseWriter, r *http.Request) {
	defer trackTime(time.Now(), "handleCount")
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
