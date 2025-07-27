package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mrshabel/chat/internal/handler"
	"github.com/mrshabel/chat/internal/util"
)

// register all the handlers with their appropriate routes
func RegisterRoutes(roomHandler *handler.RoomHandler, userHandler *handler.UserHandler) http.Handler {
	router := mux.NewRouter()

	// health check
	router.HandleFunc("/health", healthCheck)

	// websocket
	router.HandleFunc("/ws/{userId}", roomHandler.JoinRoom).Methods(http.MethodGet)

	api := router.PathPrefix("/api").Subrouter()

	// users
	users := api.PathPrefix("/users").Subrouter()
	users.HandleFunc("", userHandler.CreateUser).Methods(http.MethodPost)
	users.HandleFunc("/{id}", userHandler.GetByUserByID).Methods(http.MethodGet)

	// rooms
	rooms := api.PathPrefix("/rooms").Subrouter()
	rooms.HandleFunc("", roomHandler.CreateRoom).Methods(http.MethodPost)
	rooms.HandleFunc("", roomHandler.GetAllRooms).Methods(http.MethodGet)
	rooms.HandleFunc("/{id}", roomHandler.GetRoomByID).Methods(http.MethodGet)
	rooms.HandleFunc("/{id}/members", roomHandler.AddMember).Methods(http.MethodPost)
	rooms.HandleFunc("/{id}/members", roomHandler.GetMembers).Methods(http.MethodGet)
	rooms.HandleFunc("/{id}/members/active", roomHandler.GetActiveRoomMembers).Methods(http.MethodGet)

	// finally apply cors middleware on the router. this should be the last action performed on the router instance
	return setupCors(router)
}

func setupCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-API-KEY, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "300")

		// handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	util.WriteJSON(w, "OK", 200)
}
