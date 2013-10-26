package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	apachelog "../"
)

const (
	//format = `%h - %u %{02/Jan/2006 15:04:05 -0700}t "%m %U%q %H" %s %b %D`
	format = `%h - %u %t "%r" %s %b "%{Foobar}o"`
	addr   = "localhost:7302"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/asdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Foobar", r.URL.Query().Get("hello"))
		w.Write([]byte("hello!"))
	})

	handler := apachelog.NewHandler(format, mux, os.Stderr)
	server := &http.Server{
		Handler: handler,
		Addr:    addr,
	}
	fmt.Println("Listening on", addr)
	log.Fatal(server.ListenAndServe())
}
