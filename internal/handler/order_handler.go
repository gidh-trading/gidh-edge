package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"net/http"
)

type OrderHandler struct {
	manager *service.OrderManager
}

func NewOrderHandler(m *service.OrderManager) *OrderHandler {
	return &OrderHandler{manager: m}
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

	order, err := h.manager.SubmitOrder(r.Context(), req, uid)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendResponse(w, http.StatusOK, "success", order, "Order submitted successfully")
}

func (h *OrderHandler) GetActiveOrders(w http.ResponseWriter, r *http.Request) {
	isBacktest := r.URL.Query().Get("backtest") == "true"
	uid := r.Header.Get("X-Firebase-UID")

	orders, err := h.manager.GetActiveOrders(r.Context(), isBacktest, uid)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	h.sendResponse(w, http.StatusOK, "success", orders, "")
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
	h.sendResponse(w, code, "error", nil, msg)
}
