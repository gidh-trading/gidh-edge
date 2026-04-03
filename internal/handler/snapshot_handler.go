package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
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

// GetSnapshot handles GET /api/snapshot/{token}/{date}?interval=5m
func (h *SnapshotHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	token, _ := strconv.ParseUint(chi.URLParam(r, "token"), 10, 32)
	date, _ := time.Parse("2006-01-02", chi.URLParam(r, "date"))

	// Extract interval from query params, default to 1m
	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "1m"
	}

	snap, _ := h.service.GetFullDaySnapshot(r.Context(), uint32(token), date, interval)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.JSONResponse{Status: "success", Data: snap})
}
