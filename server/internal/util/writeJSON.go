package util

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, body any, status int, header http.Header) error {
	for key, values := range header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(body)
}
