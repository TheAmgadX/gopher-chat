package utils

import (
	"encoding/json"
	"net/http"
)

func WriteJsonErrors(w http.ResponseWriter, err string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"error": err,
	})
}

func WriteJson(w http.ResponseWriter, response map[string]any) error {
	w.Header().Set("Content-Type", "application/json")

	return json.NewEncoder(w).Encode(response)
}
