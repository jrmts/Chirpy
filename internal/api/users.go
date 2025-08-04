package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jrmts/Chrispy/internal/auth"
	"github.com/jrmts/Chrispy/internal/database"
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
	if user.Password == "" {
		respondWithError(writer, http.StatusBadRequest, "Password is required")
		return
	}
	user.HashedPassword, err = auth.HashPassword(user.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	// user.ID = uuid.New()
	// user.CreatedAt = time.Now()
	// user.UpdatedAt = user.CreatedAt

	// Save user to the database
	dbUser, err := config.Queries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
	})
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

func (config *APIConfig) LoginUser(writer http.ResponseWriter, request *http.Request) {
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
	if user.Email == "" || user.Password == "" {
		respondWithError(writer, http.StatusUnauthorized, "Email and password are required")
		return
	}
	dbUser, err := config.Queries.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		log.Printf("Failed to get user by email: %v", err)
		respondWithError(writer, http.StatusUnauthorized, "Failed to get user")
		return
	}

	// hashedPassword, _ := auth.HashPassword(user.Password)
	err = auth.CheckPasswordHash(user.Password, dbUser.HashedPassword)
	if err != nil {
		log.Printf("Invalid email or password")
		respondWithError(writer, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	respondWithJSON(writer, http.StatusOK, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})

}
