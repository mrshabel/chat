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

	"github.com/joho/godotenv"
	"github.com/mrshabel/chat/internal/config"
	"github.com/mrshabel/chat/internal/database"
	"github.com/mrshabel/chat/internal/handler"
	"github.com/mrshabel/chat/internal/repository"
	"github.com/mrshabel/chat/internal/router"
	"github.com/mrshabel/chat/internal/service"
	"github.com/mrshabel/chat/internal/service/ws"
)

var addr = flag.String("addr", "127.0.0.1:8000", "HTTP service address")

func main() {
	flag.Parse()
	godotenv.Load()

	// load configs
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// initialize db
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// initialize repositories, services
	userRepo := repository.NewUserRepository(db.DB)
	roomRepo := repository.NewRoomRepository(db.DB)
	messageRepo := repository.NewMessageRepository(db.DB)

	userService := service.NewUserService(userRepo)
	roomService := service.NewRoomService(roomRepo)
	messageService := service.NewMessageService(messageRepo)

	// start ws hub
	hub := ws.NewHub(roomService, messageService)
	go hub.Run()

	// create handlers

	roomHandler := handler.NewRoomHandler(hub, roomService, userService, messageService)
	userHandler := handler.NewUserHandler(userService)

	// register all routes
	r := router.RegisterRoutes(roomHandler, userHandler)

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
