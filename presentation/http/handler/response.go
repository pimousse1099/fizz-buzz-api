// Package handler contains the HTTP handlers and their request/response helpers.
package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes the message as a JSON string body with the given status.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, message)
}
