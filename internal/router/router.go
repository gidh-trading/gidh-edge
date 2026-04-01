package router

import (
	"gidh-edge/internal/handler"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// New creates and configures the main service router
func NewRouter(h *handler.EdgeHandler) http.Handler {
	r := chi.NewRouter()

	// --- Standard Middleware ---
	r.Use(middleware.Recoverer) // Prevent crashes on panics
	r.Use(middleware.Logger)    // Trace every request

	// CORS Configuration for Next.js UI
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- Routes ---
	r.Route("/api", func(r chi.Router) {
		// 1. Discovery
		r.Get("/instruments", h.GetInstruments)
		r.Get("/calendar", h.GetCalendar)

		// 2. Historical Rehydration
		r.Get("/history/bars", h.GetHistoryBars)
		r.Get("/history/signals", h.GetHistorySignals)

		// 3. Behavioral Baselines
		r.Get("/baselines", h.GetBaselines)

		// 4. Live Snapshot (The Stitcher)
		r.Get("/snapshot/{token}/{date}", h.GetSnapshot)

		// 6. System Health
		r.Get("/health", h.GetHealth)
		r.Get("/engine-status", h.GetEngineStatus)
	})

	return r
}
