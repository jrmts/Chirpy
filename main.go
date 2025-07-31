package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	// Importing pq for PostgreSQL driver
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/jrmts/Chrispy/internal/database"
	_ "github.com/lib/pq"
)

// type Server struct {
// 	Address string
// 	Handler http.Handler
// }

type apiConfig struct {
	fileserverHits atomic.Int32
	queries        *database.Queries
	platform       string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func badWordReplace(chirp string) string {
	splitedchirp := strings.Split(chirp, " ")
	var sliceCleanedChirp []string
	var cleanedChirp string
	for _, word := range splitedchirp {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			sliceCleanedChirp = append(sliceCleanedChirp, "****") // "****"
		} else {
			sliceCleanedChirp = append(sliceCleanedChirp, word)
		}
	}
	cleanedChirp = strings.Join(sliceCleanedChirp, " ")
	return cleanedChirp
}

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
	writer.WriteHeader(code)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(writer).Encode(data)
	if err != nil {
		log.Printf("Failed to write JSON response: %v", err)
		return
	}
}

func (config *apiConfig) createUser(writer http.ResponseWriter, request *http.Request) {
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

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	// Save user to the database
	_, err = config.queries.CreateUser(context.Background(), database.CreateUserParams{
		Email: user.Email,
		ID:    user.ID,
	})
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		respondWithError(writer, http.StatusInternalServerError, "Failed to create user")
		return
	}
	// writer.WriteHeader(http.StatusCreated)
	respondWithJSON(writer, http.StatusCreated, user)
	log.Printf("User created successfully: %v", user)
}

func (config *apiConfig) middlewareMetricsInc(realHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.fileserverHits.Add(1)
		realHandler.ServeHTTP(writer, request)
	})
}

func (config *apiConfig) handleMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	hits := config.fileserverHits.Load()
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

// * is was called validateChirp, but it was not used in the latest code
func (config *apiConfig) chirps(writer http.ResponseWriter, response *http.Request) {
	type ChirpRequest struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}
	if response.Method != http.MethodPost {
		// writer.WriteHeader(http.StatusMethodNotAllowed)
		// returnBody := ChirpRequest{
		// 	Body: "Chirp must be a POST request",
		// }
		// writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		// err := json.NewEncoder(writer).Encode(returnBody)
		// if err != nil {
		// 	log.Printf("Error writing response: %v", err)
		// 	return
		// }
		respondWithError(writer, http.StatusMethodNotAllowed, "Chirp must be a POST request")
		return
	}

	decoder := json.NewDecoder(response.Body)
	var chirpRequest ChirpRequest
	err := decoder.Decode(&chirpRequest)
	if err != nil {
		// writer.WriteHeader(http.StatusBadRequest)
		// returnBody := ChirpRequest{
		// 	Body: "Chirp must be a valid JSON object",
		// }
		// writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		// err = json.NewEncoder(writer).Encode(returnBody)
		// if err != nil {
		// 	log.Printf("Error writing response: %v", err)
		// 	return
		// }
		respondWithError(writer, http.StatusBadRequest, "Chirp must be a valid JSON object")
		return
	}

	if len(chirpRequest.Body) > 140 {
		// writer.WriteHeader(http.StatusBadRequest)
		// returnBody := ChirpRequest{
		// 	Body: "Chirp is too long.",
		// }
		// writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		// err = json.NewEncoder(writer).Encode(returnBody)
		// if err != nil {
		// 	log.Printf("Error writing response: %v", err)
		// 	return
		// }
		respondWithError(writer, http.StatusBadRequest, "Chirp is too long.")
		return
	}

	// save the chirp to the database
	userId, err := uuid.Parse(chirpRequest.UserId)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Invalid user_id format")
		return
	}

	dbChirp, err := config.queries.CreateChirp(context.Background(), database.CreateChirpParams{
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

func (config *apiConfig) resetMetric(writer http.ResponseWriter, request *http.Request) {
	// config.fileserverHits.Store(0)
	// log.Println("Metrics reset to zero")
	if config.platform != "dev" {
		respondWithError(writer, http.StatusBadRequest, "Resetting metrics is only allowed in development mode")
		return
	}
	err := config.queries.DeleteAllUsers(context.Background())
	if err != nil {
		log.Printf("Failed to reset users: %v", err)
		respondWithError(writer, http.StatusBadRequest, "Failed to reset users")
		return
	}

}

func handleHealthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

func main() {
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}
	dbQueries := database.New(db)
	apiConfiguration := &apiConfig{
		fileserverHits: atomic.Int32{},
		queries:        dbQueries,
		platform:       platform,
	}

	// const port = "8080"
	// const filepathRoot = "."
	port := flag.String("port", "8080", "TCP port to listen on")
	filepathRoot := flag.String("root", ".", "Static file root directory")
	flag.Parse()

	mux := http.NewServeMux()

	fileSystem := http.Dir(*filepathRoot) // "." means current directory
	fileServer := http.FileServer(fileSystem)
	// Serve static files from the "/app/" path
	// The StripPrefix removes the "/app/" prefix from the request URL
	// so that the file server can serve files from the root directory.
	mux.Handle("/app/", apiConfiguration.middlewareMetricsInc(http.StripPrefix(("/app/"), fileServer)))

	mux.HandleFunc("GET /api/healthz", handleHealthCheck)
	mux.HandleFunc("GET /admin/metrics", apiConfiguration.handleMetrics)
	// mux.HandleFunc("/reset", apiConfiguration.resetMetric)
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiConfiguration.resetMetric))

	mux.HandleFunc("/api/chirps", apiConfiguration.chirps)

	mux.HandleFunc("POST /api/users", apiConfiguration.createUser)

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", *port)
	log.Fatal(server.ListenAndServe())
}
