package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
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
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
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
	// if user.ExpiresAt == 0 {
	// 	user.ExpiresAt = 60 * 60 // Set default expiration to 1 hour
	// } else if user.ExpiresAt > 60*60 {
	// 	user.ExpiresAt = 60 * 60 // Ensure expiration is not more than 1 hour
	// }
	// user.ExpiresAt = time.Now().Add(1 * time.Hour)

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

	token, err := auth.MakeJWT(dbUser.ID, config.SecretKey, 1*time.Hour)
	if err != nil {
		log.Printf("Failed to create JWT: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to create JWT")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Failed to create refresh token: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to create refresh token")
		return
	}

	dbRefreshToken := database.RefreshToken{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		// RevokedAt: time.Time{}, // zero value means not revoked
	}
	_, err = config.Queries.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token:     dbRefreshToken.Token,
		UserID:    dbRefreshToken.UserID,
		CreatedAt: dbRefreshToken.CreatedAt,
		UpdatedAt: dbRefreshToken.UpdatedAt,
		ExpiresAt: dbRefreshToken.ExpiresAt,
		// RevokedAt: time.Time{},
	})
	if err != nil {
		log.Printf("Failed to create refresh token in database: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to create refresh token in database")
		return
	}

	respondWithJSON(writer, http.StatusOK, User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  dbUser.IsChirpyRed,
	})

}

func (config *APIConfig) RefreshToken(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		respondWithError(writer, http.StatusBadRequest, "Only POST method is allowed")
		return
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid or missing token")
		return
	}
	refreshToken, err := config.Queries.GetRefreshToken(context.Background(), token)
	if err != nil || refreshToken.RevokedAt.Valid {
		log.Printf("Failed to get refresh token: %v", err)
		respondWithError(writer, http.StatusUnauthorized, "Failed to get refresh token - Invalid or missing refresh token")
		return
	}
	// newRefreshToken, err := auth.MakeRefreshToken()
	// if err != nil {
	// 	log.Printf("Failed to create refresh token: %v", err)
	// 	respondWithError(writer, http.StatusInternalServerError, "Failed to create refresh token")
	// 	return
	// }
	// Update the existing refresh token in the database
	userUUID, err := config.Queries.GetUserFromRefreshToken(context.Background(), refreshToken.Token)
	if err != nil {
		log.Printf("Failed to get user from refresh token: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to get user from refresh token")
		return
	}
	// _, err = config.Queries.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
	// 	Token:     newRefreshToken,
	// 	UserID:    userUUID,
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// 	ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	// })
	// if err != nil {
	// 	log.Printf("Failed to create new refresh token in database: %v", err)
	// 	respondWithError(writer, http.StatusInternalServerError, "Failed to create new refresh token in database")
	// 	return
	// }
	// *
	// * the function below was missing and that's what caused the error
	newAccessToken, err := auth.MakeJWT(userUUID, config.SecretKey, 1*time.Hour)
	if err != nil {
		log.Printf("Failed to create new access token: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to create new access token")
		return
	}

	respondWithJSON(writer, http.StatusOK, map[string]string{
		"token": newAccessToken})
}

func (config *APIConfig) RevokeToken(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		respondWithError(writer, http.StatusBadRequest, "Only POST method is allowed")
		return
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		log.Printf("Invalid or missing token: %v", err)
		respondWithError(writer, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	// refreshTokenToRevoke, err := config.Queries.GetRefreshToken(context.Background(), token)
	// if err != nil {
	// 	log.Printf("Failed to get refresh token: %v", err)
	// 	respondWithError(writer, http.StatusUnauthorized, "Invalid or missing refresh token")
	// 	return
	// }

	err = config.Queries.RevokeRefreshToken(context.Background(), token) // refreshTokenToRevoke.Token)
	if err != nil {
		log.Printf("Failed to revoke refresh token: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to revoke refresh token")
		return
	}

	// respondWithJSON(writer, http.StatusNoContent, ) //"Refresh token revoked successfully")
	writer.WriteHeader(http.StatusNoContent)

}

func (config *APIConfig) UpdateUser(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut {
		respondWithError(writer, http.StatusMethodNotAllowed, "User update must be a PUT request")
		return
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid or missing token")
		return
	}

	_, err = auth.ValidateJWT(token, config.SecretKey)
	if err != nil {
		log.Printf("Failed to validate JWT: %v", err)
		respondWithError(writer, http.StatusUnauthorized, "Invalid token")
		return
	}

	var user User
	err = json.NewDecoder(request.Body).Decode(&user)
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
	respondWithJSON(writer, http.StatusOK, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
	log.Printf("User created successfully: %v", dbUser)
}

func (config *APIConfig) UpdateChirpyRed(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		respondWithError(writer, http.StatusMethodNotAllowed, "User update must be a POST request")
		return
	}
	polkaApiKey, err := auth.GetAPIKey(request.Header)
	if err != nil {
		log.Printf("Invalid or missing Polka API key: %v", err)
		respondWithError(writer, http.StatusUnauthorized, "Invalid or missing Polka API key")
		return
	}
	if polkaApiKey != config.PolkaKey {
		log.Printf("Invalid Polka API key: %v", polkaApiKey)
		respondWithError(writer, http.StatusUnauthorized, "Invalid Polka API key")
		return
	}

	type UpdateChirpyRedRequest struct {
		UserID uuid.UUID         `json:"user_id"`
		Event  string            `json:"event"`
		Data   map[string]string `json:"data"`
	}

	var updateRequest UpdateChirpyRedRequest
	err = json.NewDecoder(request.Body).Decode(&updateRequest)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, ok := updateRequest.Data["user_id"]
	if !ok || userID == "" {
		log.Printf("Missing user_id in request data: %v", updateRequest.Data)
		respondWithError(writer, http.StatusBadRequest, "Missing user_id in request data")
		return
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid user_id format: %v", userID)
		respondWithError(writer, http.StatusBadRequest, "Invalid user_id format")
		return
	}

	// debugging
	log.Printf("UpdateChirpyRed: Looking for user ID: %s", updateRequest.UserID)
	log.Printf("UpdateChirpyRed: Looking for user UUID from Data field: %s", userUUID)
	// if updateRequest.UserID == uuid.Nil {
	// 	log.Printf("Missing or invalid user_id in request: %v", updateRequest.UserID)
	// 	respondWithError(writer, http.StatusBadRequest, "Missing or invalid user_id")
	// 	return
	// }

	if updateRequest.Event != "user.upgraded" {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = config.Queries.GetUserById(context.Background(), userUUID)
	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	err = config.Queries.UpdateChirpyRed(context.Background(), userUUID)
	if err != nil {
		log.Printf("Failed to update Chirpy Red status: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to update Chirpy Red status")
		return
	}

	log.Printf("User %s upgraded to Chirpy Red", userUUID)
	writer.WriteHeader(http.StatusNoContent)

}
