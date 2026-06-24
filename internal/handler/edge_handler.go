package handler

import (
	"encoding/json"
	"gidh-edge/internal/models"
	"gidh-edge/internal/service"
	"io"
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

// GetAllInstruments handles /api/instruments/all (global list)
func (h *EdgeHandler) GetAllInstruments(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	queryDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		queryDate = time.Now()
	}

	instruments, err := h.service.GetAllInstruments(r.Context(), queryDate)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Instrument retrieval failed")
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

func (h *EdgeHandler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()

	resp, err := h.service.ProxyRequest(r.Context(), r.Method, uri, r.Body, r.Header)
	if err != nil {
		h.sendError(w, http.StatusBadGateway, "Backend engine is currently unavailable")
		return
	}
	defer resp.Body.Close()

	// Copy response headers from backend to client
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *EdgeHandler) GetPricePotential(w http.ResponseWriter, r *http.Request) {
	stockName := r.URL.Query().Get("stock_name")
	interval := r.URL.Query().Get("interval")

	if stockName == "" {
		h.sendError(w, http.StatusBadRequest, "Missing required query parameter: stock_name")
		return
	}

	if interval == "" {
		interval = "1m" // fallback default matching your snapshot logic
	}

	potential, err := h.service.GetPricePotential(r.Context(), stockName, interval)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve price potential metrics")
		return
	}

	if potential == nil {
		potential = []models.PricePotential{}
	}

	h.sendResponse(w, http.StatusOK, "success", potential, "")
}

func (h *EdgeHandler) GetInstrumentProfiles(w http.ResponseWriter, r *http.Request) {
	// Extract date query parameter from request
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		h.sendError(w, http.StatusBadRequest, "Missing required query parameter: date")
		return
	}

	// Pass the date down into the context or service layer
	profiles, err := h.service.GetInstrumentProfiles(r.Context(), dateStr)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve instrument metrics profiles")
		return
	}

	if profiles == nil {
		profiles = []models.InstrumentProfile{} // Explicit empty array payload
	}

	h.sendResponse(w, http.StatusOK, "success", profiles, "Instrument profiles retrieved successfully")
}

func (h *EdgeHandler) GetVWAPDistancePercentiles(w http.ResponseWriter, r *http.Request) {
	// Extract date query parameter from request
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		h.sendError(w, http.StatusBadRequest, "Missing required query parameter: date")
		return
	}

	percentiles, err := h.service.GetVWAPDistancePercentiles(r.Context(), dateStr)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to retrieve VWAP distance percentiles")
		return
	}

	if percentiles == nil {
		percentiles = []models.VWAPDistancePercentiles{} // Explicit empty array payload for frontend safety
	}

	h.sendResponse(w, http.StatusOK, "success", percentiles, "VWAP distance percentiles retrieved successfully")
}
