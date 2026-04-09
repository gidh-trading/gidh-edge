package router

import (
	"gidh-edge/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(edgeH *handler.EdgeHandler, snapH *handler.SnapshotHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/instruments", edgeH.GetAvailableInstruments)

		// Updated endpoint for the T=0 initialization
		r.Get("/snapshot/{token}/{date}", snapH.GetSnapshot)

		// Updated utility endpoint (Date removed, renamed to market-dna)
		r.Get("/market-dna/{token}/{date}", edgeH.GetMarketDNA)

		r.Post("/orders/submit", edgeH.SubmitOrder)
		r.Get("/orders/active", edgeH.GetActiveOrders)

	})
	return r
}
