package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var addr = flag.String("addr", "127.0.0.1:8000", "HTTP service address")

func main() {
	flag.Parse()
	router := mux.NewRouter()
	server := &http.Server{
		Handler: router,
		Addr:    *addr,
	}

	// start ws hub
	hub := NewHub()
	go hub.run()

	router.HandleFunc("/ws/{userId}", func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["userId"]
		if id == "" {
			http.Error(w, "User ID is required", http.StatusUnprocessableEntity)
			return
		}
		q := r.URL.Query()
		roomID := q.Get("roomId")
		if roomID == "" {
			http.Error(w, "Room ID is required", http.StatusUnprocessableEntity)
			return
		}
		username := q.Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusUnprocessableEntity)
			return
		}
		client := &Client{ID: id, RoomID: roomID, Username: username}

		serveWS(hub, client, w, r)
	})

	// start server in background
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v\n", err)
		}
	}()
	log.Println("server started successfully")

	<-ctx.Done()
	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("failed to shutdown server")
	}

	log.Println("server shutdown complete")
}
