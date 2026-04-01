package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
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

	// 1. Validation
	token, err := strconv.ParseUint(tokenStr, 10, 32)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid instrument token")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// 2. Fetch Stitched Data (DB Memory + Engine Pulse)
	snapshot, err := h.service.GetFullDaySnapshot(r.Context(), uint32(token), date)
	if err != nil {
		logger.Errorf("Failed to build snapshot for %d: %v", token, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to rehydrate session data")
		return
	}

	// 3. Success Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status: "success",
		Data:   snapshot,
	})
}

func (h *SnapshotHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  "error",
		Message: message,
	})
}
