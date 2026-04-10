package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"gidh-edge/pkg/logger"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	svc *service.OrderService // Changed from manager *service.OrderManager
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: s}
}

func (h *OrderHandler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	var req models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	uid := r.Header.Get("X-Firebase-UID")
	if uid == "" {
		h.sendError(w, http.StatusUnauthorized, "Missing user context")
		return
	}

	// Call svc instead of manager
	order, err := h.svc.SubmitOrder(r.Context(), req, uid)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendResponse(w, http.StatusOK, "success", order, "Order submitted to engine")
}

func (h *OrderHandler) GetActiveOrders(w http.ResponseWriter, r *http.Request) {
	uid := r.Header.Get("X-Firebase-UID")

	// Parse query parameters
	query := r.URL.Query()
	tokenStr := query.Get("token")
	date := query.Get("date")

	// QoL: Immediate validation for required filters
	if tokenStr == "" || date == "" {
		h.sendError(w, http.StatusBadRequest, "Missing token or date parameter")
		return
	}

	token, err := strconv.ParseUint(tokenStr, 10, 32)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid instrument token format")
		return
	}

	// Pass the filters to the service
	orders, err := h.svc.GetActiveOrders(r.Context(), uid, uint32(token), date)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	h.sendResponse(w, http.StatusOK, "success", orders, "")
}

func (h *OrderHandler) UpdateOrderRisk(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	var req struct {
		StopLoss   float64 `json:"stop_loss"`
		TakeProfit float64 `json:"take_profit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.svc.UpdateRisk(r.Context(), orderID, req.StopLoss, req.TakeProfit)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendResponse(w, http.StatusOK, "success", nil, "Risk updated")
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	err := h.svc.CancelOrder(r.Context(), orderID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.sendResponse(w, http.StatusOK, "success", nil, "Order cancelled")
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
	logger.Errorf("Error: %+v", msg)
	h.sendResponse(w, code, "error", nil, msg)
}
