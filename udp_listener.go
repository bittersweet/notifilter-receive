package main

import (
	"encoding/json"
	"github.com/hoisie/redis"
	"log"
	"net"
)

const maxPacketSize = 1024 * 1024

type Stat struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func trackStat(stat Stat) {
	var client redis.Client
	var key = stat.Key
	client.Set(key, []byte(stat.Value))
	val, _ := client.Get(stat.Key)
	log.Println(key, string(val))
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

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
	}
}
