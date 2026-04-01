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

func (h *EdgeHandler) GetAvailableInstruments(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")

	// Default to today if no date provided
	queryDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		queryDate = time.Now()
	}

	instruments, err := h.service.GetInstruments(r.Context(), queryDate)
	if err != nil {
		h.respond(w, http.StatusInternalServerError, "error", nil, "Discovery failed")
		return
	}

	h.respond(w, http.StatusOK, "success", instruments, "")
}

func (h *EdgeHandler) respond(w http.ResponseWriter, code int, status string, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	})
}
