package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/jackc/pgx/v5/pgxpool"

    "github.com/mayamengya/myceli/internal/middleware"
)

type UserHandler struct {
    DB *pgxpool.Pool
}

// No firebase_uid here: identity comes from the verified token, never the body.
type registerRequest struct {
    DisplayName string `json:"display_name"`
}

// Register creates (or updates) the caller's row in our users table.
// Called after the user's first Firebase login. Requires VerifyToken middleware.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    firebaseUID, ok := r.Context().Value(middleware.ContextKeyFirebaseUID).(string)
    if !ok || firebaseUID == "" {
        // VerifyToken didn't run: a wiring bug in main.go, not a client error.
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
        return
    }

    var req registerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
        return
    }

    if req.DisplayName == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "display_name required"})
        return
    }

    // Upsert so retries are safe. DO UPDATE (not DO NOTHING) so RETURNING
    // always yields the id, even when the row already exists.
    var userID string
    err := h.DB.QueryRow(r.Context(),
        `INSERT INTO users (firebase_uid, display_name)
         VALUES ($1, $2)
         ON CONFLICT (firebase_uid) DO UPDATE SET display_name = EXCLUDED.display_name
         RETURNING id`,
        firebaseUID,
        req.DisplayName,
    ).Scan(&userID)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "failed to register user"})
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"id": userID})
}
