package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
)

const maxPacketSize = 1024 * 1024

var db *sqlx.DB

type Stat struct {
	Key   string         `json:"key"`
	Value types.JsonText `json:"value"`
}

type Notifier struct {
	Id               int            `db:"id"`
	NotificationType string         `db:"notification_type"`
	Class            string         `db:"class"`
	Template         string         `db:"template"`
	Rules            types.JsonText `db:"rules"`
	// type slack/email/direct to phone
	// email address, slack channel, phone number, how to store?
}

type Incoming struct {
	Id         int       `db:"id"`
	Class      string    `db:"class"`
	ReceivedAt time.Time `db:"received_at"`
	Data       string    `db:"data"`
}

func (n *Notifier) getRules() []*Rule {
	rules := []*Rule{}
	n.Rules.Unmarshal(&rules)
	return rules
}

func (n *Notifier) checkRules(s *Stat) bool {
	rules := n.getRules()
	rules_met := true
	for _, rule := range rules {
		if !rule.Met(s) {
			fmt.Printf("Rule not met -- Key: %s, Type: %s, Setting %s, Value %s\n", rule.Key, rule.Type, rule.Setting, rule.Value)
			rules_met = false
		}
	}
	if !rules_met {
		fmt.Printf("Stopping notification of id: %d, rules not met\n", n.Id)
		return false
	}

	return true
}

func (i *Incoming) FormattedData() string {
	return string(i.Data)
}

func (s *Stat) toMap() map[string]interface{} {
	m := map[string]interface{}{}
	s.Value.Unmarshal(&m)
	return m
}

func (s *Stat) persist() {
	var incomingId int
	query := `INSERT INTO incoming(received_at, class, data) VALUES($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, time.Now(), s.Key, s.Value).Scan(&incomingId)
	if err != nil {
		log.Fatal("persist()", err)
	}
	fmt.Printf("class: %s id: %d\n", s.Key, incomingId)
}

func (s *Stat) notify() {
	notifiers := []Notifier{}
	err := db.Select(&notifiers, "SELECT * FROM notifiers WHERE class=$1", s.Key)
	if err != nil {
		log.Fatal("db.Select ", err)
	}
	fmt.Printf("Found %d notifiers\n", len(notifiers))

Notify:
	for i := 0; i < len(notifiers); i++ {
		notifier := notifiers[i]
		nt := notifier.NotificationType
		fmt.Printf("Notifying notifier id: %d type: %s\n", notifier.Id, nt)
		fmt.Printf("Incoming data: %v\n", s.toMap())

		if !notifier.checkRules(s) {
			continue Notify
		}

		if nt == "email" {
			sendEmailNotification(s, &notifier)
		} else if nt == "slack" {
			sendSlackNotification(s, &notifier)
		}
	}
}

func countRows() int {
	var rows int
	err := db.QueryRow("select count(*) from incoming").Scan(&rows)
	if err != nil {
		log.Fatal("rowcount: ", err)
	}

	return rows
}

func listenToUDP(conn *net.UDPConn) {
	buffer := make([]byte, maxPacketSize)
	for {
		bytes, err := conn.Read(buffer)
		if err != nil {
			log.Println("UDP read error: ", err.Error())
			continue
		}

		msg := make([]byte, bytes)
		copy(msg, buffer)

		var stat Stat
		err = json.Unmarshal(msg, &stat)
		if err != nil {
			log.Println(err)
			log.Printf("%+v\n", stat)
		}

		stat.persist()
		stat.notify()
	}
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		log.Fatal("ResolveUDPAddr", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("ListenUDP", err)
	}

	go listenToUDP(conn)
	http.HandleFunc("/favicon.ico", handleFavicon)
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/new", handleNew)
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/v1/count", handleCount)

	db, err = sqlx.Connect("postgres", "user=markmulder dbname=notifier sslmode=disable")
	if err != nil {
		log.Fatal("DB Open()", err)
	}
	defer db.Close()

	rows := countRows()
	fmt.Println("Total rows:", rows)

	fmt.Println("Will start listening on port 8000")
	http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe ", err)
	}
}
