package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrshabel/chat/internal/handler"
	"github.com/mrshabel/chat/internal/router"
	"github.com/mrshabel/chat/internal/service/ws"
)

var addr = flag.String("addr", "127.0.0.1:8000", "HTTP service address")

func main() {
	flag.Parse()

	// start ws hub
	hub := ws.NewHub()
	go hub.Run()

	// create handlers
	roomHandler := handler.NewRoomHandler(hub)

	// register all routes
	r := router.RegisterRoutes(roomHandler)

	// http server
	server := &http.Server{
		Handler: r,
		Addr:    *addr,
	}

	// start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v\n", err)
		}
	}()
	log.Println("server started successfully")

	if err := cleanup(server); err != nil {
		log.Fatal(err)
	}
	log.Println("server shutdown complete")
}

func cleanup(server *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}
