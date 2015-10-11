package main

import (
	"log"
	"net"
	"encoding/json"
)

const maxPacketSize = 1024 * 1024

type Stat struct {
  Key   string `json:"key"`
  Value string `json:"value"`
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
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("UDP read error: ", err.Error())
			continue
		}

		msg := make([]byte, n)
		copy(msg, buffer)

		log.Println(string(msg))

		var stat Stat
    err = json.Unmarshal(msg, &stat)
    if err == nil {
      log.Printf("%+v\n", stat)
    } else {
      log.Println(err)
      log.Printf("%+v\n", stat)
    }
	}
}
