package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"gidh-edge/pkg/logger"
	"io"
	"net/http"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

// HandleProxy forwards the incoming request verbatim to the backend engine
func (h *OrderHandler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	// RequestURI captures the full path + query parameters (e.g., /api/order/cancel?order_id=xyz)
	uri := r.URL.RequestURI()

	resp, err := h.service.ProxyOrderRequest(r.Context(), r.Method, uri, r.Body, r.Header)
	if err != nil {
		logger.Errorf("Failed to proxy order request to engine: %v", err)
		h.sendError(w, http.StatusBadGateway, "Trading engine is currently unavailable")
		return
	}
	defer resp.Body.Close()

	// Copy headers from the engine's response back to the UI
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Write the matching status code and pipe the raw JSON response payload
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *OrderHandler) sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  "error",
		Message: msg,
	})
}
