package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"net/http"
	"time"
)

type EdgeHandler struct {
	service *service.EdgeService
}

func NewEdgeHandler(s *service.EdgeService) *EdgeHandler {
	return &EdgeHandler{service: s}
}

// GetAvailableInstruments handles GET /api/instruments?date=YYYY-MM-DD
func (h *EdgeHandler) GetAvailableInstruments(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")

	queryDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// Default to current date if missing or invalid
		queryDate = time.Now()
	}

	instruments, err := h.service.GetAvailable(r.Context(), queryDate)
	if err != nil {
		h.sendResponse(w, http.StatusInternalServerError, "error", nil, "Instrument discovery failed")
		return
	}

	h.sendResponse(w, http.StatusOK, "success", instruments, "")
}

func (h *EdgeHandler) sendResponse(w http.ResponseWriter, code int, status string, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	})
}
