package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"net/http"
)

type EdgeHandler struct {
	svc *service.EdgeService
}

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

func (h *EdgeHandler) GetBaselines(w http.ResponseWriter, r *http.Request) {
	// ... Logic to extract token/date and call svc.GetBaselines ...
}
