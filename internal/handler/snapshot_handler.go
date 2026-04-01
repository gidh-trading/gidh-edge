package handler

import (
	"encoding/json"
	"gidh-edge/internal/service"
	"gidh-edge/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type SnapshotHandler struct {
	service *service.SnapshotService
}

func NewSnapshotHandler(s *service.SnapshotService) *SnapshotHandler {
	return &SnapshotHandler{service: s}
}

func (h *SnapshotHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")
	dateStr := chi.URLParam(r, "date")

	// 1. Validate Token
	token, err := strconv.ParseUint(tokenStr, 10, 32)
	if err != nil {
		logger.Warnf("Invalid token format: %s", tokenStr)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// 2. Validate Date (YYYY-MM-DD) as per Specification
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		logger.Warnf("Invalid date format: %s. Expected YYYY-MM-DD", dateStr)
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// 3. Call Service Layer
	snapshot, err := h.service.GetFullDaySnapshot(r.Context(), uint32(token), date)
	if err != nil {
		logger.Errorf("Service error fetching snapshot: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 4. Encode Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}
