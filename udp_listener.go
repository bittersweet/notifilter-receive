package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	// "github.com/hoisie/redis"
	// _ "github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
	"log"
	"net"
	"time"
)

const maxPacketSize = 1024 * 1024

var db *sql.DB

type Stat struct {
	Key   string         `json:"key"`
	Value types.JsonText `json:"value"`
}

func (s *Stat) persist() {
	fmt.Println(s.Key)
	fmt.Println(s.Value)
	var incomingId int
	// err := db.QueryRow(`INSERT INTO incoming(received_at, data) VALUES(?, ?)`, time.Now(), s.Value).Scan(&incomingId)
	err := db.QueryRow(`INSERT INTO incoming(received_at, data) VALUES($1, $2) RETURNING id`, time.Now(), s.Value).Scan(&incomingId)
	if err != nil {
		log.Fatal("persist()", err)
	}
	fmt.Println("saved to id: ", incomingId)
}

func trackStat(stat Stat) {
	// var client redis.Client
	log.Printf("%+v\n", stat.Key)
	log.Printf("%+v\n", stat.Value)
	// client.Lpush("marktest", []byte(stat.Value))
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

	db, err = sql.Open("postgres", "user=markmulder dbname=notifier sslmode=disable")
	if err != nil {
		log.Fatal("DB Open()", err)
	}
	defer db.Close()

	var rows int
	err = db.QueryRow("select count(*) from incoming").Scan(&rows)
	if err != nil {
		log.Fatal("rowcount: ", err)
	}
	fmt.Println("== rows: ", rows)

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

		trackStat(stat)
		stat.persist()
	}

	fmt.Println("Listening from here")
}
