// internal/handler/order_handler.go

package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"io"
	"net/http"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) HandleOrderPlace(w http.ResponseWriter, r *http.Request) {
	// Forward the entire request (method, body, headers) to the service
	resp, err := h.service.ProxyOrder(r.Context(), r.Method, r.Body, r.Header)
	if err != nil {
		h.sendError(w, http.StatusBadGateway, "Trading engine is currently unavailable")
		return
	}
	defer resp.Body.Close()

	// Copy response headers from backend to client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Forward the status code and body back to the user
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *OrderHandler) sendResponse(w http.ResponseWriter, code int, status string, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  status,
		Data:    data,
		Message: msg,
	})
}

func (h *OrderHandler) sendError(w http.ResponseWriter, code int, msg string) {
	// Reusing your existing JSON error response format if needed for local errors
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// (Standard JSONResponse encoding here...)
}
