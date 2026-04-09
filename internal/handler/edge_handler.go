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

type EdgeHandler struct {
	service *service.EdgeService
}

func NewEdgeHandler(s *service.EdgeService) *EdgeHandler {
	return &EdgeHandler{service: s}
}

func (h *EdgeHandler) GetAvailableInstruments(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	queryDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		queryDate = time.Now()
	}

	instruments, err := h.service.GetInstruments(r.Context(), queryDate)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Instrument discovery failed")
		return
	}
	h.sendResponse(w, http.StatusOK, "success", instruments, "")
}

func (h *EdgeHandler) GetMarketDNA(w http.ResponseWriter, r *http.Request) {
	tokenStr := chi.URLParam(r, "token")
	dateStr := chi.URLParam(r, "date")

	token, err := strconv.ParseUint(tokenStr, 10, 32)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid token")
		return
	}

	queryDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	dna, err := h.service.GetMarketDNA(r.Context(), uint32(token), queryDate)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Market DNA not found")
		return
	}
	h.sendResponse(w, http.StatusOK, "success", dna, "")
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

func (h *EdgeHandler) sendError(w http.ResponseWriter, code int, msg string) {
	h.sendResponse(w, code, "error", nil, msg)
}
