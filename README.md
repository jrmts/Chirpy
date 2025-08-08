# Chirpy

Chirpy is a minimalist Twitter-like microblogging platform built in Go. It was developed as part of my hands-on learning journey through the [Boot.dev backend Go track](https://boot.dev).

This project helped me apply concepts such as:

- HTTP routing and handler design
- Middleware for metrics and authentication
- Secure password hashing and JWT-based login
- SQL database integration with `sqlc`
- REST API design with JSON request/response handling
- Basic testing and code organization in Go

### Learning Roadmap
Chirpy was my first serious Go backend project—tools like JWT, SQL migrations, file serving, and middleware deepened my understanding. I’m proud of the progress and eager to build on this foundation. This project represents a snapshot of my learning—thank you for checking it out!

### Topic to dive deeper
- Cryptography
- Postgres (SQL)
- Error handling
- Unit tests


---

##  Built With

- **Go** – Leveraging core packages like `net/http`, `sqlc`, `github.com/golang-jwt/jwt/v5`, and more  
- **PostgreSQL** – Persistent storage for users and posts  


### Features

- User signup and login with hashed passwords
- JWT-based authentication
- Protected endpoints (like posting chirps)
- Static file serving
- Metrics and health check endpoints
- Admin-only reset functionality 


### Prerequisites

- Go 1.23+
- PostgreSQL (or `psql`)
- `sqlc` CLI  
- `.env` file containing your `DB_URL` and `PLATFORM=dev` for local setup

### Installation

```bash
git clone https://github.com/jrmts/Chirpy.git
cd Chirpy
go install github.com/kyleconroy/sqlc/cmd/sqlc@latest
sqlc generate
go run main.go
```

### Endpoints
- POST /api/users – Create users
- POST /api/login – Authenticate and get JWT token
- POST /api/chirps – Create chirps (authorized)
- GET /api/chirps – List all chirps
- GET /api/healthz, /admin/metrics, /admin/reset – Admin and health utilities

Find more details in the internal/api packages and route definitions in main.go.















