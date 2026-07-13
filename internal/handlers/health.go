package handlers

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
    DB *pgxpool.Pool
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
    // Verify the DB is reachable.
    if err := h.DB.Ping(context.Background()); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error":  err.Error(),
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}