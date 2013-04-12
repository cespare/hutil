package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cespare/hutil"
)

const (
	addr = "localhost:6677"
	statsAddr = "localhost:6678"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte("Hello!"))
	})

	stat := hutil.NewStatRecorder(mux)
	//mux.Handle("/stats", stat.HandlerFunc())
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
