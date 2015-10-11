package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/bittersweet/notifilter/elasticsearch"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
)

const maxPacketSize = 1024 * 1024

var db *sqlx.DB

// Event will hold incoming data and will be persisted to ES eventually
type Event struct {
	Application string `json:"application"`
	Identifier  string `json:"identifier"`
	requestID   string
	Data        types.JsonText `json:"data"`
}

// dataToMap transforms the raw JSON data into a map
func (e *Event) dataToMap() map[string]interface{} {
	m := map[string]interface{}{}
	err := e.Data.Unmarshal(&m)
	if err != nil {
		e.log("Error in dataToMap():", err)
		return map[string]interface{}{}
	}
	return m
}

// persist saves the incoming event to Elasticsearch
func (e *Event) persist() {
	err := elasticsearch.Persist(e.requestID, e.Application, e.Identifier, e.dataToMap())
	if err != nil {
		e.log("Error persisting to ElasticSearch:", err)
	}
}

// notify checks to see if we have notifiers set up for this event and if the
// rules for those notifications have been satisfied
func (e *Event) notify() {
	notifiers := []Notifier{}
	err := db.Select(&notifiers, "SELECT * FROM notifiers WHERE application=$1 AND event_name=$2", e.Application, e.Identifier)
	if err != nil {
		log.Fatal("db.Select ", err)
	}
	e.log("[NOTIFY] found %d notifiers", len(notifiers))

	for i := 0; i < len(notifiers); i++ {
		notifier := notifiers[i]
		notifier.notify(e, notifier.newNotifier())
	}
}

func (e *Event) log(msg string, args ...interface{}) {
	logStr := fmt.Sprintf(msg, args...)
	log.Printf("%s %s\n", e.requestID, logStr)
}

// incomingItems creates a channel that we can place events on so the main loop
// can keep listening to incoming events
func incomingItems() chan<- []byte {
	incomingChan := make(chan []byte)

	// Open a channel with a capacity of 10.000 events
	// This will only block the sender if the buffer fills up.
	// If we do not buffer any event that gets sent to the channel will be
	// dropped if we can not handle it.
	tasks := make(chan Event, 10000)

	// Generate unique ID to tag requests
	// Thanks to https://blog.cloudflare.com/go-at-cloudflare/
	idGenerator := make(chan string)
	go func() {
		h := sha1.New()
		c := []byte(time.Now().String())
		for {
			h.Write(c)
			idGenerator <- fmt.Sprintf("%x", h.Sum(nil))
		}
	}()

	// Use 4 workers that will concurrently grab Events of the channel and
	// persist+notify
	for i := 0; i < 4; i++ {
		go func() {
			for event := range tasks {
				event.persist()
				event.notify()
			}
		}()
	}

	go func() {
		for {
			select {
			case b := <-incomingChan:
				var event Event
				err := json.Unmarshal(b, &event)
				if err != nil {
					log.Println(err)
				}
				requestID := <-idGenerator
				event.requestID = requestID
				tasks <- event
			}
		}
	}()

	fmt.Println("incomingItems launched")
	return incomingChan
}

// listenToUDP opens a UDP connection that we will listen on
func listenToUDP(conn *net.UDPConn) {
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

// main handles setting up connections for UDP/TCP and connecting to Postgres
// Note, sqlx uses a connection pool internally.
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

	db, err = sqlx.Connect("postgres", "user=markmulder dbname=notifilter_development sslmode=disable")
	if err != nil {
		log.Fatal("DB Open()", err)
	}
	defer db.Close()

	http.HandleFunc("/v1/count", handleCount)

	fmt.Println("Will start listening on port 8000")
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe ", err)
	}
}
