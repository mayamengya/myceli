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
    // Don't just return 200 — actually verify the DB is reachable.
    // A health check that doesn't check dependencies is a lie.
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