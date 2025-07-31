package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func (config *APIConfig) CreateUser(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		respondWithError(writer, http.StatusBadRequest, "Only POST method is allowed")
		return
	}
	var user User
	err := json.NewDecoder(request.Body).Decode(&user)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid request body")
		return
	}
	if user.Email == "" {
		respondWithError(writer, http.StatusBadRequest, "Email is required")
		return
	}

	// user.ID = uuid.New()
	// user.CreatedAt = time.Now()
	// user.UpdatedAt = user.CreatedAt

	// Save user to the database
	dbUser, err := config.Queries.CreateUser(context.Background(), user.Email)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to create user")
		return
	}
	// writer.WriteHeader(http.StatusCreated)
	respondWithJSON(writer, http.StatusCreated, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
	log.Printf("User created successfully: %v", dbUser)
}
