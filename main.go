package main

import (
	"flag"
	"log"
	"net/http"
	"online-class/internal/signal"
)

func main() {
	port := flag.String("port", "8080", "http server port")
	flag.Parse()

	hub := signal.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		signal.ServeWs(hub, w, r)
	})

	log.Printf("Signaling server running on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
