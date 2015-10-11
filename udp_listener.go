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
	"gopkg.in/fatih/set.v0"
)

const maxPacketSize = 1024 * 1024

var db *sqlx.DB
var keys = set.New()

type Stat struct {
	Key   string         `json:"key"`
	Value types.JsonText `json:"value"`
}

type Incoming struct {
	Id         int       `db:"id"`
	Class      string    `db:"class"`
	ReceivedAt time.Time `db:"received_at"`
	Data       string    `db:"data"`
}

func (i *Incoming) FormattedData() string {
	return string(i.Data)
}

func (i *Incoming) keys() []string {
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(i.Data), &parsed)
	if err != nil {
		log.Fatal("json.Unmarshal", err)
	}

	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	fmt.Println("Keys:", keys)
	return keys
}

func (i *Incoming) toMap() map[string]interface{} {
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(i.Data), &parsed)
	if err != nil {
		log.Fatal("json.Unmarshal", err)
	}

	return parsed
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
	fmt.Printf("Incoming data: %v\n", s.toMap())
	fmt.Printf("Found %d notifiers\n", len(notifiers))

	for i := 0; i < len(notifiers); i++ {
		notifier := notifiers[i]
		notifier.notify(s, notifier.newNotifier())
	}
}

func (s *Stat) keys() []string {
	sMap := s.toMap()
	keys := make([]string, 0, len(sMap))
	for k := range sMap {
		keys = append(keys, k)
	}
	return keys
}

func countRows() int {
	var rows int
	err := db.QueryRow("SELECT COUNT(*) FROM incoming").Scan(&rows)
	if err != nil {
		log.Fatal("rowcount: ", err)
	}

	return rows
}

func listenToUDP(conn *net.UDPConn) {
	updates := keyLogger()

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
		updates <- stat.keys()
		stat.notify()
	}
}

func keyLogger() chan<- []string {
	updates := make(chan []string)

	go func() {
		for {
			select {
			case incomingKeys := <-updates:
				// iterate over slice and append if not found
				for _, k := range incomingKeys {
					keys.Add(k)
				}
				fmt.Println("Keys:", keys)
			}
		}
	}()

	fmt.Println("keyLogger launched")
	return updates
}

func updateKeysFromLatest() {
	incoming := []Incoming{}
	err := db.Select(&incoming, "SELECT * FROM incoming ORDER BY id DESC LIMIT 10")
	if err != nil {
		log.Fatal("db.Select incoming ", err)
	}

	for _, record := range incoming {
		fmt.Printf("%#v\n", record.Data)
		rKeys := record.keys()

		for _, k := range rKeys {
			fmt.Println("Adding:", k)
			keys.Add(k)
		}
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
	http.HandleFunc("/preview", handlePreview)
	http.HandleFunc("/v1/count", handleCount)
	http.HandleFunc("/static/", staticHandler)

	db, err = sqlx.Connect("postgres", "user=markmulder dbname=notifier sslmode=disable")
	if err != nil {
		log.Fatal("DB Open()", err)
	}
	defer db.Close()

	rows := countRows()
	fmt.Println("Total rows:", rows)

	updateKeysFromLatest()

	fmt.Println("Will start listening on port 8000")
	http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe ", err)
	}
}
