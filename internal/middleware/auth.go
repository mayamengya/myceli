package middleware

import (
    "context"
    "encoding/json"
    "net/http"
    "strings"

    "github.com/mayamengya/myceli/internal/auth"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Context keys (typed to avoid collisions with other middleware)
type contextKey string

const (
    ContextKeyUserID      contextKey = "userID"      // your internal UUID
    ContextKeyFirebaseUID contextKey = "firebaseUID"  // Firebase's UID
)

type AuthMiddleware struct {
    Firebase *auth.Client
    DB       *pgxpool.Pool
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            writeUnauthorized(w, "missing authorization header")
            return
        }
        idToken := strings.TrimPrefix(authHeader, "Bearer ")

        // Verify with Firebase
        firebaseUID, err := m.Firebase.VerifyToken(r.Context(), idToken)
        if err != nil {
            writeUnauthorized(w, "invalid token")
            return
        }

        // Look up internal user ID from users table
        var userID string
        err = m.DB.QueryRow(r.Context(),
            "SELECT id FROM users WHERE firebase_uid = $1",
            firebaseUID,
        ).Scan(&userID)

        if err != nil {
            // User authenticated with Firebase but doesn't exist in DB yet.
            // This happens on first login (handle registration separately)
            writeUnauthorized(w, "user not registered")
            return
        }

        // Attach to context so handlers can access it
        ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
        ctx = context.WithValue(ctx, ContextKeyFirebaseUID, firebaseUID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func writeUnauthorized(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}