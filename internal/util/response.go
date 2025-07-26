package util

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, data any, status int) {
	// set appropriate headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, message string, status int) {
	WriteJSON(w, map[string]string{"message": message}, status)
}
