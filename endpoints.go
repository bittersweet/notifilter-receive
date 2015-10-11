package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx/types"
)

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Println("== inside handleindex!!!")
	defer trackTime(time.Now(), "handleIndex")

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

func handleCreateRule(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	defer trackTime(time.Now(), "handleCreateRule")

	err := r.ParseForm()
	if err != nil {
		log.Fatal("handleCreateRule", err)
	}
	fmt.Println("incoming parameters")
	fmt.Printf("%v\n", r.Form)

	notification_type := r.Form.Get("notification_type")
	class := r.Form.Get("class")
	template := r.Form.Get("template")
	rules := r.Form.Get("rules")

	_, err = db.NamedExec(`INSERT INTO notifiers (notification_type, class, template, rules) VALUES (:notification_type, :class, :template, :rules)`,
		map[string]interface{}{
			"notification_type": notification_type,
			"class":             class,
			"template":          template,
			"rules":             types.JsonText(rules),
		})

	if err != nil {
		log.Fatal("insert named query", err)
	}
}

func handleCount(w http.ResponseWriter, r *http.Request) {
	defer trackTime(time.Now(), "handleCount")

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
