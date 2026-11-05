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

// VerifyToken checks the Firebase ID token in the Authorization header and
// stores the verified Firebase UID in context. It does not require the user
// to exist in our DB, so it can protect /users/register.
func (m *AuthMiddleware) VerifyToken(next http.Handler) http.Handler {
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

        ctx := context.WithValue(r.Context(), ContextKeyFirebaseUID, firebaseUID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// RequireUser looks up the internal users.id for the verified Firebase UID
// and stores it in context. Must run after VerifyToken.
func (m *AuthMiddleware) RequireUser(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        firebaseUID, ok := r.Context().Value(ContextKeyFirebaseUID).(string)
        if !ok || firebaseUID == "" {
            // VerifyToken didn't run: a wiring bug in main.go, not a client error.
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
            return
        }

        // Look up internal user ID from users table
        var userID string
        err := m.DB.QueryRow(r.Context(),
            "SELECT id FROM users WHERE firebase_uid = $1",
            firebaseUID,
        ).Scan(&userID)

        if err != nil {
            // User authenticated with Firebase but hasn't called /users/register yet.
            writeUnauthorized(w, "user not registered")
            return
        }

        ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func writeUnauthorized(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}
