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

func (h *SnapshotHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")
	dateStr := chi.URLParam(r, "date")

	token, err := strconv.ParseUint(tokenStr, 10, 32)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid token")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "1m"
	}

	snap, err := h.service.GetFullDaySnapshot(r.Context(), uint32(token), date, interval)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to generate snapshot")
		return
	}

	h.sendResponse(w, http.StatusOK, "success", snap, "")
}

func (h *SnapshotHandler) sendResponse(w http.ResponseWriter, code int, status string, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	})
}

func (h *SnapshotHandler) sendError(w http.ResponseWriter, code int, msg string) {
	h.sendResponse(w, code, "error", nil, msg)
}
