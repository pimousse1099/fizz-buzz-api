// Package handler contains the HTTP handlers and their shared response helpers.
// The standard library offers no JSON response helper (only http.Error for
// plain text) and no constant for the JSON content type, so both live here.
package handler

import (
	"encoding/json"
	"net/http"
)

const contentTypeJSON = "application/json"

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes the message as a JSON string body with the given status.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, message)
}
