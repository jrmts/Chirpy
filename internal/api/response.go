package api

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(writer http.ResponseWriter, code int, msg string) {
	writer.WriteHeader(code)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := map[string]string{"error": msg}
	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

func respondWithJSON(writer http.ResponseWriter, code int, data interface{}) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(code)
	err := json.NewEncoder(writer).Encode(data)
	if err != nil {
		log.Printf("Failed to write JSON response: %v", err)
		return
	}
}
