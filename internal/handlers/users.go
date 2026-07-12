package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
    DB *pgxpool.Pool
}

type registerRequest struct {
    FirebaseUID string `json:"firebase_uid"`
    DisplayName string `json:"display_name"`
}

// Register creates a user in our DB after Firebase auth.
// Called once after the user's first Firebase login.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
        return
    }

    if req.FirebaseUID == "" || req.DisplayName == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "firebase_uid and display_name required"})
        return
    }

    var userID string
    err := h.DB.QueryRow(r.Context(),
        `INSERT INTO users (firebase_uid, display_name)
         VALUES ($1, $2)
         ON CONFLICT (firebase_uid) DO UPDATE SET display_name = EXCLUDED.display_name
         RETURNING id`,
        req.FirebaseUID,
        req.DisplayName,
    ).Scan(&userID)

    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "failed to register user"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"id": userID})
}