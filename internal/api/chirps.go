package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jrmts/Chrispy/internal/auth"
	"github.com/jrmts/Chrispy/internal/database"
)

// chirps handles the creation of a new chirp.
//
//	is was called validateChirp, but it was not used in the latest code
func (config *APIConfig) Chirps(writer http.ResponseWriter, request *http.Request) {
	type ChirpRequest struct {
		Body string `json:"body"`
		// UserId string `json:"user_id"`
	}
	if request.Method != http.MethodPost {
		respondWithError(writer, http.StatusMethodNotAllowed, "Chirp must be a POST request")
		return
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	userID, err := auth.ValidateJWT(token, config.SecretKey)
	if err != nil {
		log.Printf("Failed to vallidate token: %v", err)
		respondWithError(writer, http.StatusUnauthorized, "Invalid token")
		return
	}

	decoder := json.NewDecoder(request.Body)
	var chirpRequest ChirpRequest
	err = decoder.Decode(&chirpRequest)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Chirp must be a valid JSON object")
		return
	}

	if len(chirpRequest.Body) > 140 {
		respondWithError(writer, http.StatusBadRequest, "Chirp is too long.")
		return
	}

	// save the chirp to the database
	// userId, err := uuid.Parse(chirpRequest.UserId)
	// if err != nil {
	// 	respondWithError(writer, http.StatusBadRequest, "Invalid user_id format")
	// 	return
	// }
	if userID != uuid.Nil {
		_, err = config.Queries.GetUserById(context.Background(), userID)
		if err != nil {
			respondWithError(writer, http.StatusBadRequest, "User does not exist")
			return
		}
	}

	_, err = config.Queries.GetUserById(context.Background(), userID)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "User does not exist")
		return
	}

	dbChirp, err := config.Queries.CreateChirp(context.Background(), database.CreateChirpParams{
		UserID: userID,
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

// GetChirpByID retrieves a chirp by its ID.
func (config *APIConfig) GetChirpByID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		respondWithError(writer, http.StatusMethodNotAllowed, "Must be a GET request")
		return
	}
	chirpID := request.PathValue("id")
	if chirpID == "" {
		respondWithError(writer, http.StatusBadRequest, "Chirp ID required")
		return
	}
	id, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid Chirp ID format")
		return
	}
	dbChirp, err := config.Queries.GetChirpByID(context.Background(), id)
	if err != nil {
		log.Printf("Failed to get chirp by ID: %v", err)
		respondWithError(writer, http.StatusNotFound, fmt.Sprintf("Failed to get chirp by ID: %v", err))
		return
	}
	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
	}
	log.Printf("Chirp recieved successfully: %v", chirp)
	respondWithJSON(writer, http.StatusOK, chirp)
}

func (config *APIConfig) DeleteOneChirp(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		respondWithError(writer, http.StatusMethodNotAllowed, "Chirp must be a DELETE request")
		return
	}

	requestUserToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	requestUserUUID, err := auth.ValidateJWT(requestUserToken, config.SecretKey)
	if err != nil {
		log.Printf("Failed to validate JWT: %v", err)
		respondWithError(writer, http.StatusUnauthorized, "Invalid token")
		return
	}

	id := request.PathValue("id")
	if id == "" {
		respondWithError(writer, http.StatusBadRequest, "Chirp ID required")
		return
	}
	chirpToDeleteID, err := uuid.Parse(id)
	// if chirpToDeleteID == uuid.Nil {
	// 	respondWithError(writer, http.StatusBadRequest, "Invalid Chirp ID format")
	// 	return
	// }
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Failed to parse Chirp ID")
		return
	}

	// var chirp Chirp
	chirp, err := config.Queries.GetChirpByID(context.Background(), chirpToDeleteID)
	if err != nil {
		log.Printf("Failed to get chirp by ID: %v", err)
		respondWithError(writer, http.StatusNotFound, fmt.Sprintf("Chirp not found: %v", err))
		return
	}

	userChirpOwner, err := config.Queries.GetUserById(context.Background(), chirp.UserID)
	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		respondWithError(writer, http.StatusNotFound, fmt.Sprintf("User not found: %v", err))
		return
	}
	userRequestOwner, err := config.Queries.GetUserById(context.Background(), requestUserUUID)
	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		respondWithError(writer, http.StatusNotFound, fmt.Sprintf("User not found: %v", err))
		return
	}

	// Check if the user making the request is the owner of the chirp
	if userRequestOwner.ID != userChirpOwner.ID {
		log.Printf("User %v is not authorized to delete chirp %v", userRequestOwner.ID, chirpToDeleteID)
		respondWithError(writer, http.StatusForbidden, "You are not authorized to delete this chirp")
		return
	}

	// chirpOwnerUUID := chirp.UserID
	// // if chirpOwnerUUID == uuid.Nil {
	// // 	log.Printf("Chirp %v does not have a valid user ID", chirpToDeleteID)
	// // 	respondWithError(writer, http.StatusNotFound, "Chirp does not have a valid user ID")
	// // 	return
	// // }

	// userChirpOwner, err := config.Queries.GetUserById(context.Background(), chirpOwnerUUID)
	// if err != nil {
	// 	log.Printf("Failed to get user by ID: %v", err)
	// 	respondWithError(writer, http.StatusNotFound, fmt.Sprintf("User not found: %v", err))
	// 	return
	// }

	// if chirpOwnerUUID != userChirpOwner.ID || requestUserUUID != userChirpOwner.ID {
	// 	log.Printf("User %v is not authorized to delete chirp %v", requestUserUUID, chirpToDeleteID)
	// 	respondWithError(writer, http.StatusForbidden, "You are not authorized to delete this chirp")
	// 	return
	// }
	err = config.Queries.DeleteOneChirps(context.Background(), chirpToDeleteID)
	if err != nil {
		log.Printf("Failed to delete chirp: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to delete chirp.")
		return
	}
	log.Printf("Chirp %v deleted successfully by user %v", chirpToDeleteID, requestUserUUID)
	writer.WriteHeader(http.StatusNoContent) // No content response

}
