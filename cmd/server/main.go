package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/joho/godotenv"

    "github.com/mayamengya/myceli/internal/db"
    "github.com/mayamengya/myceli/internal/handlers"
	"github.com/mayamengya/myceli/internal/auth"
    custommiddleware "github.com/mayamengya/myceli/internal/middleware"
)

func main() {
    // Load .env file. In production this would be real env vars 
    // godotenv is a dev convenience, not a production dependency.
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, reading from environment")
    }

    ctx := context.Background()

    // Initialize DB pool. Fail fast if unreachable.
    pool, err := db.NewPool(ctx)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer pool.Close()
    log.Println("Database connection pool established")
	
	firebaseClient, err := auth.NewClient(ctx) 
	if err != nil { 
		log.Fatalf("Failed to initialize Firebase: %v", err) 
		} 
	log.Println("Firebase auth client initialized")

    // Wire up handlers
    healthHandler := &handlers.HealthHandler{DB: pool}
	userHandler := &handlers.UserHandler{DB: pool}
    authMiddleware := &custommiddleware.AuthMiddleware{
    Firebase: firebaseClient,
    DB:       pool,
}

    // Router
    r := chi.NewRouter()
    r.Use(middleware.Logger)    // logs every request
    r.Use(middleware.Recoverer) // catches panics, returns 500
	r.Get("/health", healthHandler.Check)
	r.Post("/users/register", userHandler.Register) // public — no auth required
    r.Get("/health", healthHandler.Check)
	// Protected routes - auth middleware applied
	r.Group(func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    // TODO: future endpoints go here
	})

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s", port)
    if err := http.ListenAndServe(":"+port, r); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}