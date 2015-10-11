package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
)

const maxPacketSize = 1024 * 1024

var db *sqlx.DB

type Event struct {
	Key   string         `json:"key"`
	Value types.JsonText `json:"value"`
}

type Incoming struct {
	Id         int       `db:"id"`
	Class      string    `db:"class"`
	ReceivedAt time.Time `db:"received_at"`
	Data       string    `db:"data"`
}

type ESPayload struct {
	Key        string                 `json:"key"`
	ReceivedAt time.Time              `json:"received_at"`
	Data       map[string]interface{} `json:"data"`
}

func (i *Incoming) FormattedData() string {
	return string(i.Data)
}

func (i *Incoming) toMap() map[string]interface{} {
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(i.Data), &parsed)
	if err != nil {
		log.Fatal("json.Unmarshal", err)
	}

	return parsed
}

func (e *Event) toMap() map[string]interface{} {
	m := map[string]interface{}{}
	e.Value.Unmarshal(&m)
	return m
}

func (e *Event) persist() {
	var incomingID int
	query := `INSERT INTO incoming(received_at, class, data) VALUES($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, time.Now(), e.Key, e.Value).Scan(&incomingID)
	if err != nil {
		log.Fatal("persist()", err)
	}
	fmt.Printf("class: %s id: %d\n", e.Key, incomingID)
}

func (e *Event) notify() {
	notifiers := []Notifier{}
	err := db.Select(&notifiers, "SELECT * FROM notifiers WHERE class=$1", e.Key)
	if err != nil {
		log.Fatal("db.Select ", err)
	}
	fmt.Printf("Incoming data: %v\n", e.toMap())
	fmt.Printf("Found %d notifiers\n", len(notifiers))

	go func() {
		payload := ESPayload{
			Key:        e.Key,
			ReceivedAt: time.Now(),
			Data:       e.toMap(),
		}
		body, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:9200/notifilter/event/?pretty", "application/json", bytes.NewReader(body))
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Print(string(body))
	}()

	for i := 0; i < len(notifiers); i++ {
		notifier := notifiers[i]
		notifier.notify(e, notifier.newNotifier())
	}
}

func countRows() int {
	var rows int
	err := db.QueryRow("SELECT COUNT(*) FROM incoming").Scan(&rows)
	if err != nil {
		log.Fatal("rowcount: ", err)
	}

	return rows
}

func incomingItems() chan<- []byte {
	incomingChan := make(chan []byte)

	go func() {
		var i = 0
		for {
			select {
			case b := <-incomingChan:
				i = i + 1
				fmt.Println(i)

				var Event Event
				err := json.Unmarshal(b, &Event)
				if err != nil {
					log.Println(err)
					log.Printf("%+v\n", Event)
				}
				Event.persist()
				Event.notify()
			}
		}
	}()

	fmt.Println("incomingItems launched")
	return incomingChan
}

func listenToUDP(conn *net.UDPConn) {
	// updates := keyLogger()
	incomingChan := incomingItems()

	buffer := make([]byte, maxPacketSize)
	for {
		bytes, err := conn.Read(buffer)
		if err != nil {
			log.Println("UDP read error: ", err.Error())
			continue
		}

		msg := make([]byte, bytes)
		copy(msg, buffer)
		incomingChan <- msg
	}
}

func main() {
	runtime.GOMAXPROCS(4)

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

	fmt.Println("Will start listening on port 8000")
	http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe ", err)
	}
}
