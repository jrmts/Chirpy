package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jrmts/Chrispy/internal/database"
)

// chirps handles the creation of a new chirp.
//
//	is was called validateChirp, but it was not used in the latest code
func (config *APIConfig) Chirps(writer http.ResponseWriter, request *http.Request) {
	type ChirpRequest struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}
	if request.Method != http.MethodPost {
		respondWithError(writer, http.StatusMethodNotAllowed, "Chirp must be a POST request")
		return
	}

	decoder := json.NewDecoder(request.Body)
	var chirpRequest ChirpRequest
	err := decoder.Decode(&chirpRequest)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Chirp must be a valid JSON object")
		return
	}

	if len(chirpRequest.Body) > 140 {
		respondWithError(writer, http.StatusBadRequest, "Chirp is too long.")
		return
	}

	// save the chirp to the database
	userId, err := uuid.Parse(chirpRequest.UserId)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid user_id format")
		return
	}
	if userId != uuid.Nil {
		_, err = config.Queries.GetUserById(context.Background(), userId)
		if err != nil {
			respondWithError(writer, http.StatusBadRequest, "User does not exist")
			return
		}
	}

	_, err = config.Queries.GetUserById(context.Background(), userId)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "User does not exist")
		return
	}

	dbChirp, err := config.Queries.CreateChirp(context.Background(), database.CreateChirpParams{
		UserID: userId,
		Body:   badWordReplace(chirpRequest.Body), //chirpRequest.Body,
	})
	if err != nil {
		log.Printf("Failed to create chirp: %v", err)
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Failed to create chirp %v", err))
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
	}

	log.Printf("Chirp created successfully: %v", chirp)
	respondWithJSON(writer, http.StatusCreated, chirp)
}

// get chirps retrieves all chirps for a user.
func (config *APIConfig) GetChirps(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		respondWithError(writer, http.StatusMethodNotAllowed, "Chirps must be a GET request")
		return
	}

	dbChirps, err := config.Queries.GetAllChirps(context.Background())
	if err != nil {
		log.Printf("Failed to get chirps: %v", err)
		respondWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Failed to get chirps: %v", err))
		return
	}
	// Convert sliceOfChirps to a slice of Chirp structs
	var chirps []Chirp
	for _, dbChirp := range dbChirps {
		chirp := Chirp{
			ID:        dbChirp.ID,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
		}
		chirps = append(chirps, chirp)
	}
	log.Printf("Chirps retrieved successfully: %v", chirps)
	respondWithJSON(writer, http.StatusOK, chirps)
}
