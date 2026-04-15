package jsonwriter

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Wrap responses in an object for flexibility
type listEnvelope[T any] struct {
	Data []T `json:"data"`
}

// Fancy errors
type errorEnvelope struct {
	Error string `json:"error"`
}

func WriteJSONArray[T any](w http.ResponseWriter, status int, items []T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(listEnvelope[T]{Data: items})
}

func WriteJSONErr(w http.ResponseWriter, status int, msg string, err error) {
	fmt.Println(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error: msg,
	})
}
