// internal/handler/order_handler.go

package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	service    *service.OrderService
	edgeServ   *service.EdgeService
	isBacktest bool
}

func NewOrderHandler(s *service.OrderService, es *service.EdgeService, mode string) *OrderHandler {
	return &OrderHandler{
		service:    s,
		edgeServ:   es,
		isBacktest: mode == "backtest",
	}
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

// HandleGetHistoricalOrders processes GET /api/orders/{date}
func (h *OrderHandler) HandleGetHistoricalOrders(w http.ResponseWriter, r *http.Request) {
	dateParam := chi.URLParam(r, "date")
	parsedDate, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	orders, err := h.service.GetHistoricalOrders(r.Context(), parsedDate)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch historical orders")
		return
	}

	if orders == nil {
		orders = []models.OrderBookEntry{} // Return empty array [] to keep frontend happy
	}

	h.sendResponse(w, http.StatusOK, "success", orders, "Historical orders retrieved successfully")
}

// HandleGetHistoricalPositions processes GET /api/positions/history/{date}
func (h *OrderHandler) HandleGetHistoricalPositions(w http.ResponseWriter, r *http.Request) {
	dateParam := chi.URLParam(r, "date")
	parsedDate, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	// 1. Try to proxy to the backend
	resp, err := h.service.ProxyHistoricalPositions(r.Context(), r.Method, dateParam, r.Body, r.Header)

	// 2. If proxy succeeds, stream it
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	// 3. Fallback: If proxy fails or backend is offline, fetch from local DB
	positions, dbErr := h.service.GetHistoricalPositions(r.Context(), parsedDate)
	if dbErr != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to fetch positions from engine and local database")
		return
	}

	if positions == nil {
		positions = []models.Position{}
	}

	h.sendResponse(w, http.StatusOK, "success", positions, "Retrieved local historical positions (Engine offline)")
}

// HandleVirtualContractNote processes GET /api/orders/vcn natively from the Edge layer
func (h *OrderHandler) HandleVirtualContractNote(w http.ResponseWriter, r *http.Request) {

	if h.isBacktest {
		resp, err := h.edgeServ.FetchGlobalBacktestVCN(r.Context())
		if err != nil {
			h.sendError(w, http.StatusBadGateway, "Simulation execution engine is currently unreachable")
			return
		}
		defer resp.Body.Close()

		// Stream the backend's single account metrics directly to the user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		return
	}

	note, err := h.service.GetVirtualContractNote(r.Context())
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to construct contract note ledger: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(note)
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
