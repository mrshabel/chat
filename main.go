package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", "127.0.0.1:8000", "HTTP service address")

func main() {
	flag.Parse()
	// start ws hub
	hub := newHub()
	go hub.run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("failed to start server: %v\n", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	// invalid path or method
	if r.URL.Path != "/" {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}
