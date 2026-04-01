package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"net/http"
	"strconv"
	"time"
)

type EdgeHandler struct {
	svc *service.EdgeService
}

func NewEdgeHandler(svc *service.EdgeService) *EdgeHandler {
	return &EdgeHandler{svc: svc}
}

// sendJSON provides a standard JSON response wrapper
func (h *EdgeHandler) sendJSON(w http.ResponseWriter, status int, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := models.JSONResponse{Status: "success", Data: data, Message: msg}
	if status >= 400 {
		resp.Status = "error"
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *EdgeHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	dates, err := h.svc.GetCalendar(r.Context())
	if err != nil {
		h.sendJSON(w, 500, nil, "Failed to fetch calendar")
		return
	}
	h.sendJSON(w, 200, dates, "")
}

func (h *EdgeHandler) GetInstruments(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.sendJSON(w, 400, nil, "Invalid date format. Use YYYY-MM-DD")
		return
	}
	data, err := h.svc.GetInstruments(r.Context(), date)
	if err != nil {
		h.sendJSON(w, 500, nil, err.Error())
		return
	}
	h.sendJSON(w, 200, data, "")
}

func (h *EdgeHandler) GetHistoryBars(w http.ResponseWriter, r *http.Request) {
	token, _ := strconv.ParseUint(r.URL.Query().Get("token"), 10, 32)
	date, _ := time.Parse("2006-01-02", r.URL.Query().Get("date"))

	data, err := h.svc.GetHistoryBars(r.Context(), uint32(token), date)
	if err != nil {
		h.sendJSON(w, 500, nil, "History retrieval failed")
		return
	}
	h.sendJSON(w, 200, data, "")
}

func (h *EdgeHandler) GetBaselines(w http.ResponseWriter, r *http.Request) {
	token, _ := strconv.ParseUint(r.URL.Query().Get("token"), 10, 32)
	date, _ := time.Parse("2006-01-02", r.URL.Query().Get("date"))

	data, err := h.svc.GetBaselines(r.Context(), uint32(token), date)
	if err != nil {
		h.sendJSON(w, 500, nil, "Baseline retrieval failed")
		return
	}
	h.sendJSON(w, 200, data, "")
}

func (h *EdgeHandler) GetEngineStatus(w http.ResponseWriter, r *http.Request) {
	status := h.svc.GetEngineStatus(r.Context())
	h.sendJSON(w, 200, status, "")
}
