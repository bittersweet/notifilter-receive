package main

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/bittersweet/notifilter/elasticsearch"
)

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

func handleCount(es elasticsearch.ElasticsearchClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer trackTime(time.Now(), "handleCount")

		count, err := es.EventCount()
		if err != nil {
			log.Println("Error while getting count from ES", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jmap := map[string]int{
			"count": count,
		}

		output, err := json.MarshalIndent(jmap, "", "  ")
		if err != nil {
			log.Println("Error in /v1/count MarshalIndent", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(output)
	})
}

func handleStatistics(t time.Time) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := new(runtime.MemStats)
		runtime.ReadMemStats(m)

		systemStatus := struct {
			Uptime       int64
			NumGoroutine int    //
			Alloc        uint64 // bytes allocated and not yet freed
			TotalAlloc   uint64 // bytes allocated (even if freed)
			Sys          uint64 // bytes obtained from system
		}{
			Uptime:       time.Now().Unix() - t.Unix(),
			NumGoroutine: runtime.NumGoroutine(),
			Alloc:        m.Alloc,
			TotalAlloc:   m.TotalAlloc,
			Sys:          m.Sys,
		}

		output, err := json.MarshalIndent(systemStatus, "", "  ")
		if err != nil {
			log.Println("Error in /v1/statistics MarshalIndent", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(output)
	})
}
