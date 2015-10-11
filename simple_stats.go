package main

import (
	"fmt"
	"github.com/hoisie/redis"
	"net/http"
)

func main() {
  http.HandleFunc("/", RootHandler)
  http.ListenAndServe(":8888", nil)
	fmt.Printf("Done\n")
}

func RootHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("Done with work!\n")
	var client redis.Client
	var key = "hello"
	client.Set(key, []byte("world"))
	val, _ := client.Get("hello")
	println(key, string(val))
}
