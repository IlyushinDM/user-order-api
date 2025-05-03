package utils

import (
	"encoding/json"
	"net/http"
)

// any or interface{}, password hashing and comparison using bcrypt
func Message(status bool, message string) map[string]any {
	return map[string]any{"status": status, "message": message}
}

func Respond(w http.ResponseWriter, data map[string]any) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
