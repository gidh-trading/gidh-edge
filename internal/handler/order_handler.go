// internal/handler/order_handler.go

package handler

import (
	"context"
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

// --- Order Management Handlers ---

func (h *OrderHandler) HandleOrderPlace(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.service.ProxyOrder)
}

func (h *OrderHandler) HandleOrderModify(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.service.ProxyOrderModify)
}

func (h *OrderHandler) HandleOrderCancel(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.service.ProxyOrderCancel)
}

// --- Position Management Handlers ---

func (h *OrderHandler) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.service.ProxyPositions)
}

func (h *OrderHandler) HandlePositionMetadata(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.service.ProxyPositionMetadata)
}

func (h *OrderHandler) HandlePositionExit(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, h.service.ProxyPositionExit)
}

// --- Private Proxy Helper ---

// proxyRequest is a generic helper that handles the common logic of forwarding
// a request to the service and copying the response back to the client.
func (h *OrderHandler) proxyRequest(
	w http.ResponseWriter,
	r *http.Request,
	proxyFunc func(context.Context, string, io.Reader, http.Header) (*http.Response, error),
) {
	resp, err := proxyFunc(r.Context(), r.Method, r.Body, r.Header)
	if err != nil {
		h.sendError(w, http.StatusBadGateway, "Trading engine is currently unavailable")
		return
	}
	defer resp.Body.Close()

	// Copy all response headers from backend to client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// Forward the status code and stream the body back to the user
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// --- Response Helpers ---

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.JSONResponse{
		Status:  "error",
		Message: msg,
	})
}
