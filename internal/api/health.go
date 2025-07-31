package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func HandleHealthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

func (config *APIConfig) ResetMetric(writer http.ResponseWriter, request *http.Request) {
	// config.fileserverHits.Store(0)
	// log.Println("Metrics reset to zero")
	if config.Platform != "dev" {
		respondWithError(writer, http.StatusBadRequest, "Resetting metrics is only allowed in development mode")
		return
	}
	err := config.Queries.DeleteAllUsers(context.Background())
	if err != nil {
		log.Printf("Failed to reset users: %v", err)
		respondWithError(writer, http.StatusBadRequest, "Failed to reset users")
		return
	}

	err = config.Queries.DeleteAllChirps(context.Background())
	if err != nil {
		log.Printf("Failed to reset chirps: %v", err)
		respondWithError(writer, http.StatusBadRequest, "Failed to reset chirps")
		return
	}
}

func (config *APIConfig) HandleMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	hits := config.FileserverHits.Load()
	html := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>`, hits)
	_, err := writer.Write([]byte(html))
	if err != nil {
		log.Printf("Error writing metrics response: %v", err)
	}
}
