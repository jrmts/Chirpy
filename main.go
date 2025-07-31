package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/jrmts/Chrispy/internal/api"

	// Importing pq for PostgreSQL driver

	"github.com/joho/godotenv"
	"github.com/jrmts/Chrispy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}
	dbQueries := database.New(db)
	apiConfiguration := &api.APIConfig{
		FileserverHits: atomic.Int32{},
		Queries:        dbQueries,
		Platform:       platform,
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
	mux.Handle("/app/", apiConfiguration.MiddlewareMetricsInc(http.StripPrefix(("/app/"), fileServer)))

	mux.HandleFunc("GET /api/healthz", api.HandleHealthCheck)
	mux.HandleFunc("GET /admin/metrics", apiConfiguration.HandleMetrics)
	// mux.HandleFunc("/reset", apiConfiguration.resetMetric)
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiConfiguration.ResetMetric))

	mux.HandleFunc("POST /api/chirps", apiConfiguration.Chirps)
	mux.HandleFunc("GET /api/chirps", apiConfiguration.GetChirps)

	mux.HandleFunc("POST /api/users", apiConfiguration.CreateUser)

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", *port)
	log.Fatal(server.ListenAndServe())
}
