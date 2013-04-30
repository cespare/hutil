package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cespare/hutil/stats"
)

const (
	addr = "localhost:6677"
	statsAddr = "localhost:6678"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/bad", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte("Hello!"))
	})

	stat := stats.New(mux)
	server := &http.Server{
		Handler: stat,
		Addr: addr,
	}

	statServer := &http.Server{
		Handler: stat.HandlerFunc(),
		Addr: statsAddr,
	}
	go func() {
		log.Println("Stats listening on", statsAddr)
		log.Fatal(statServer.ListenAndServe())
	}()

	log.Println("Now listening on", addr)
	log.Fatal(server.ListenAndServe())
}
