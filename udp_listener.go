package main

import (
	// "database/sql"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"text/template"
	"time"
)

const maxPacketSize = 1024 * 1024

var db *sqlx.DB

type Stat struct {
	Key   string         `json:"key"`
	Value types.JsonText `json:"value"`
}

type dbNotifier struct {
	Id       int    `db:"id"`
	Class    string `db:"class"`
	Template string `db:"template"`
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
	notifiers := []dbNotifier{}
	err := db.Select(&notifiers, "SELECT * FROM notifiers WHERE class=$1", s.Key)
	if err != nil {
		log.Fatal("db.Select ", err)
	}
	fmt.Printf("Found %d notifiers\n", len(notifiers))
	for i := 0; i < len(notifiers); i++ {
		s.specialNotify(&notifiers[i])
	}
}

func (s *Stat) specialNotify(notifier *dbNotifier) {
	fmt.Printf("Notifying notifier id: %d\n", notifier.Id)
	var err error
	var doc bytes.Buffer

	t := template.New("notificationTemplate")
	t, err = t.Parse(notifier.Template)
	if err != nil {
		log.Fatal("t.Parse of n.Template", err)
	}

	m := map[string]interface{}{}
	s.Value.Unmarshal(&m)

	err = t.Execute(&doc, m)
	if err != nil {
		log.Fatal("t.Execute ", err)
	}

	sendEmail(s.Key, doc.Bytes())
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

const emailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

<html>
<body>
{{.Body}}
</body>
</html>`

type EmailData struct {
	From    string
	To      string
	Subject string
	Body    string
}

func sendEmail(class string, data []byte) {
	var err error
	var doc bytes.Buffer

	t := template.New("emailTemplate")
	t, err = t.Parse(emailTemplate)
	if err != nil {
		log.Fatal("t.Parse ", err)
	}
	context := &EmailData{
		"Springest Dev <developers@springest.nl>",
		"recipient@example.com",
		"Email subject line",
		string(data),
	}
	err = t.Execute(&doc, context)
	if err != nil {
		log.Fatal("t.Execute ", err)
	}

	auth := smtp.PlainAuth("", "", "", "localhost:1025")
	err = smtp.SendMail("localhost:1025", auth, "test@example.com", []string{"recipient@example.com"}, doc.Bytes())
	if err != nil {
		log.Fatal("smtp.SendMail ", err)
	}

	// drop e-mail job on a rate limited (max workers) queue
	// already experienced a connection reset by peer locally
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
	http.HandleFunc("/", handleCount)

	db, err = sqlx.Connect("postgres", "user=markmulder dbname=notifier sslmode=disable")
	if err != nil {
		log.Fatal("DB Open()", err)
	}
	defer db.Close()

	rows := countRows()
	fmt.Println("Total rows: ", rows)

	fmt.Println("Will start listening on port 8000")
	http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe ", err)
	}

}
