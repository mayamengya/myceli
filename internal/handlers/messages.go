package handlers

import (
	"encoding/json"
	"net/http"

	appMiddleware "github.com/YOURUSERNAME/myceli/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageHandler struct {
	DB *pgxpool.Pool
}

type createMessageRequest struct {
	RecipientID  string  `json:"recipient_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Content      string  `json:"content"`
	RadiusMeters int     `json:"radius_meters"`
	IsPublic     bool    `json:"is_public"`
}

func (h *MessageHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Decode request body
	var req createMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	// 2. Validate
	if req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "content is required"})
		return
	}

	if req.RadiusMeters == 0 {
		req.RadiusMeters = 50 // default radius
	}

	// 3. Get user ID from context (set by auth middleware)
	userID := r.Context().Value(appMiddleware.ContextKeyUserID).(string)

	// 4. Handle recipient — nil for public messages
	var recipientID *string
	if !req.IsPublic && req.RecipientID != "" {
		recipientID = &req.RecipientID
	}

	// 5. Insert into database
	var messageID string
	err := h.DB.QueryRow(r.Context(),
		`INSERT INTO messages 
			(sender_id, recipient_id, content, location, radius_meters, is_public)
		VALUES 
			($1, $2, $3, ST_MakePoint($4, $5)::geography, $6, $7)
		RETURNING id`,
		userID,
		recipientID,
		req.Content,
		req.Longitude, // ST_MakePoint takes longitude first, then latitude
		req.Latitude,
		req.RadiusMeters,
		req.IsPublic,
	).Scan(&messageID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create message"})
		return
	}

	// 6. Return the new message ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": messageID})
}