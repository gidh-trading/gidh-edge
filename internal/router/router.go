package router

import (
	"context"
	"gidh-edge/internal/handler"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(
	edgeH *handler.EdgeHandler,
	snapH *handler.SnapshotHandler,
	backtestH *handler.BacktestHandler,
	orderH *handler.OrderHandler,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Firebase-UID", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/available-dates", backtestH.GetAvailableDates)
		r.Get("/instruments", edgeH.GetAvailableInstruments)
		r.Get("/instruments/all", edgeH.GetAllInstruments)
		r.Get("/instruments/profiles", edgeH.GetInstrumentProfiles)
		r.Get("/instruments/vwap-percentiles", edgeH.GetVWAPDistancePercentiles)
		r.Get("/price-potential", edgeH.GetPricePotential)
		r.Get("/snapshot/{token}/{date}", snapH.GetSnapshot)
		r.Get("/market-dna/{token}/{date}", edgeH.GetMarketDNA)
		r.Get("/alerts", edgeH.HandleProxy)

		// --- Backtest Proxy Routes ---
		r.With(TimeoutMiddleware(3*time.Minute)).Post("/backtest/start", backtestH.HandleProxy)
		r.Get("/backtest/stop", backtestH.HandleProxy)
		r.Get("/backtest/available-dates", backtestH.HandleProxy)
		r.Get("/backtest/status", backtestH.HandleProxy)
		r.Post("/backtest/speed", backtestH.HandleProxy)

		// --- Order Management ---
		r.Post("/orders/place", orderH.HandleOrderPlace)
		r.Post("/orders/modify", orderH.HandleOrderModify)
		r.Post("/orders/cancel", orderH.HandleOrderCancel)
		r.Get("/orders/{date}", orderH.HandleGetHistoricalOrders)
		r.Get("/orders/vcn", orderH.HandleVirtualContractNote)

		// --- Position Management ---
		r.Get("/positions", orderH.HandleGetPositions) // From previous step
		r.Post("/positions/metadata", orderH.HandlePositionMetadata)
		r.Post("/positions/exit", orderH.HandlePositionExit)
		r.Get("/positions/history/{date}", orderH.HandleGetHistoricalPositions)
	})
	return r
}

// TimeoutMiddleware is a custom middleware that applies a timeout to a specific handler
func TimeoutMiddleware(timeout time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a context with timeout derived from the request context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// Handler completed within timeout
				return
			case <-ctx.Done():
				// Timeout occurred
				w.WriteHeader(http.StatusGatewayTimeout)
				w.Write([]byte(`{"error": "request timeout"}`))
				return
			}
		})
	}
}
