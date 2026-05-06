package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"gidh-edge/pkg/logger"
	"io"
	"net/http"
)

type BacktestHandler struct {
	service *service.BacktestService
}

func NewBacktestHandler(s *service.BacktestService) *BacktestHandler {
	return &BacktestHandler{service: s}
}

func (h *BacktestHandler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()

	resp, err := h.service.ProxyBacktestRequest(r.Context(), r.Method, uri, r.Body, r.Header)
	if err != nil {
		logger.Errorf("Failed to proxy backtest request: %v", err)
		h.sendError(w, http.StatusBadGateway, "Backtest engine is currently unavailable")
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *BacktestHandler) GetAvailableDates(w http.ResponseWriter, r *http.Request) {
	dates, err := h.service.GetAvailableDates(r.Context())
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Note: UI expects { "dates": [...] }
	h.sendResponse(w, http.StatusOK, "success", map[string]interface{}{"dates": dates}, "")
}

func (h *BacktestHandler) sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  "error",
		Message: msg,
	})
}

func (h *BacktestHandler) sendResponse(w http.ResponseWriter, code int, status string, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	})
}
