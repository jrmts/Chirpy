package api

import (
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jrmts/Chrispy/internal/database"
)

type APIConfig struct {
	FileserverHits atomic.Int32
	Queries        *database.Queries
	Platform       string
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
