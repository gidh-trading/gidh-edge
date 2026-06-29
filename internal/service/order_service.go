package service

import (
	"context"
	"fmt"
	"gidh-edge/internal/client"
	"gidh-edge/internal/models"
	"gidh-edge/internal/repo"
	"gidh-edge/pkg/logger"
	"io"
	"math"
	"net/http"
	"time"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type OrderService struct {
	engine     *client.HTTPEngineClient
	repo       *repo.OrderRepository
	kiteClient *kiteconnect.Client
}

func NewOrderService(e *client.HTTPEngineClient, repo *repo.OrderRepository, apiKey, accessToken string) *OrderService {
	var kc *kiteconnect.Client
	if apiKey != "" && accessToken != "" {
		kc = kiteconnect.New(apiKey)
		kc.SetAccessToken(accessToken)
	}

	return &OrderService{
		engine:     e,
		repo:       repo,
		kiteClient: kc,
	}
}

func (s *OrderService) GetVirtualContractNote(ctx context.Context, tradingDate string) (models.VirtualContractNoteResponse, error) {

	todayStr := time.Now().Format("2006-01-02")

	// --- 1. HISTORICAL OR SIMULATION MODE: FETCH DIRECTLY FROM DB ---
	// If a specific date is requested, or if we don't have a live client context, let edge calculate it
	isHistorical := tradingDate != "" && tradingDate != todayStr

	if isHistorical || s.kiteClient == nil {
		return s.calculateSimulationContractNote(ctx, tradingDate)
	}

	// --- 2. LIVE RETRIEVAL VIA BROKER INTERFACE (Current Session) ---
	liveTrades, err := s.kiteClient.GetTrades()
	if err != nil {
		logger.Errorf("Failed to retrieve live day trades from Zerodha: %v", err)
		return s.calculateSimulationContractNote(ctx, "") // Safe fallback to today's local table state
	}

	var response models.VirtualContractNoteResponse
	if len(liveTrades) == 0 {
		response.Trades = []models.EnrichedMockTrade{}
		return response, nil
	}

	orderParamsList := make([]kiteconnect.OrderChargesParam, len(liveTrades))
	for i, t := range liveTrades {
		prod := t.Product
		if prod == "" {
			prod = "MIS"
		}

		orderParamsList[i] = kiteconnect.OrderChargesParam{
			OrderID:         t.OrderID,
			Exchange:        t.Exchange,
			Tradingsymbol:   t.TradingSymbol,
			TransactionType: t.TransactionType,
			Variety:         "regular",
			Product:         prod,
			OrderType:       "MARKET",
			Quantity:        float64(t.Quantity),
			AveragePrice:    t.AveragePrice,
		}
	}

	chargesResponse, err := s.kiteClient.GetOrderCharges(kiteconnect.GetChargesParams{OrderParams: orderParamsList})
	if err != nil {
		logger.Errorf("Zerodha Charges Engine pipeline exception: %v", err)
		return models.VirtualContractNoteResponse{}, fmt.Errorf("failed to sync calculations from broker: %w", err)
	}

	response.Trades = make([]models.EnrichedMockTrade, len(liveTrades))
	for i, chargeDetail := range chargesResponse {
		response.Summary.Brokerage += chargeDetail.Charges.Brokerage
		response.Summary.STT += chargeDetail.Charges.TransactionTax
		response.Summary.StampDuty += chargeDetail.Charges.StampDuty
		response.Summary.ExchangeTurnoverCharge += chargeDetail.Charges.ExchangeTurnoverCharge
		response.Summary.SEBITurnoverCharge += chargeDetail.Charges.SEBITurnoverCharge
		response.Summary.GST += chargeDetail.Charges.GST.Total
		response.Summary.TotalCharges += chargeDetail.Charges.Total

		response.Trades[i] = models.EnrichedMockTrade{
			Timestamp:       liveTrades[i].FillTimestamp.Time,
			Side:            orderParamsList[i].TransactionType,
			Symbol:          orderParamsList[i].Tradingsymbol,
			Exchange:        orderParamsList[i].Exchange,
			Quantity:        int(orderParamsList[i].Quantity),
			AveragePrice:    orderParamsList[i].AveragePrice,
			AllocatedCharge: chargeDetail.Charges.Total,
		}
	}

	return s.roundSummaryMetrics(response), nil
}

func (s *OrderService) calculateSimulationContractNote(ctx context.Context, tradingDate string) (models.VirtualContractNoteResponse, error) {
	userEmail := "algo.trader@gidh.tech"

	dbPayload, err := s.repo.FetchVCNPayload(ctx, userEmail, tradingDate)
	if err != nil {
		return models.VirtualContractNoteResponse{}, fmt.Errorf("database retrieval failed for contract note: %w", err)
	}

	var response models.VirtualContractNoteResponse
	response.Summary.Brokerage = dbPayload.Summary.Brokerage
	response.Summary.STT = dbPayload.Summary.STT
	response.Summary.StampDuty = dbPayload.Summary.StampDuty // ✅ Fixed tracking assignment
	response.Summary.ExchangeTurnoverCharge = dbPayload.Summary.ExchangeTurnoverCharge
	response.Summary.SEBITurnoverCharge = dbPayload.Summary.SebiTurnoverCharge
	response.Summary.GST = dbPayload.Summary.GST
	response.Summary.TotalCharges = dbPayload.Summary.TotalCharges

	response.Trades = make([]models.EnrichedMockTrade, len(dbPayload.Trades))
	for i, dbTrade := range dbPayload.Trades {
		response.Trades[i] = models.EnrichedMockTrade{
			Timestamp:       dbTrade.Timestamp,
			Side:            dbTrade.Side,
			Symbol:          dbTrade.Symbol,
			Exchange:        "NSE",
			Quantity:        dbTrade.Quantity,
			AveragePrice:    dbTrade.Price,
			AllocatedCharge: dbTrade.Charges,
		}
	}

	return s.roundSummaryMetrics(response), nil
}

func (s *OrderService) roundSummaryMetrics(res models.VirtualContractNoteResponse) models.VirtualContractNoteResponse {
	res.Summary.Brokerage = math.Round(res.Summary.Brokerage*100) / 100
	res.Summary.STT = math.Round(res.Summary.STT*100) / 100
	res.Summary.StampDuty = math.Round(res.Summary.StampDuty*100) / 100
	res.Summary.ExchangeTurnoverCharge = math.Round(res.Summary.ExchangeTurnoverCharge*100) / 100
	res.Summary.SEBITurnoverCharge = math.Round(res.Summary.SEBITurnoverCharge*100) / 100
	res.Summary.GST = math.Round(res.Summary.GST*100) / 100
	res.Summary.TotalCharges = math.Round(res.Summary.TotalCharges*100) / 100
	return res
}

// GetHistoricalOrders proxies the call down to the repository layer
func (s *OrderService) GetHistoricalOrders(ctx context.Context, tradingDate time.Time) ([]models.OrderBookEntry, error) {
	return s.repo.GetHistoricalOrders(ctx, tradingDate)
}

func (s *OrderService) GetHistoricalPositions(ctx context.Context, tradingDate time.Time) ([]models.Position, error) {
	return s.repo.GetHistoricalPositions(ctx, tradingDate)
}

// --- Proxy Order System calls remain completely unchanged below ---

func (s *OrderService) ProxyHistoricalPositions(ctx context.Context, method string, date string, body io.Reader, headers http.Header) (*http.Response, error) {
	uri := fmt.Sprintf("/api/positions/history/%s", date)
	return s.engine.ForwardRawRequest(ctx, method, uri, body, headers)
}

func (s *OrderService) ProxyOrder(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/orders/place", body, headers)
}

func (s *OrderService) ProxyPositions(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/positions", body, headers)
}

func (s *OrderService) ProxyOrderModify(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/orders/modify", body, headers)
}

func (s *OrderService) ProxyOrderCancel(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/orders/cancel", body, headers)
}

func (s *OrderService) ProxyPositionMetadata(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/positions/metadata", body, headers)
}

func (s *OrderService) ProxyPositionExit(ctx context.Context, method string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, "/api/positions/exit", body, headers)
}
