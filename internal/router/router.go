package router

import (
	"gidh-edge/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(edgeH *handler.EdgeHandler, snapH *handler.SnapshotHandler, orderH *handler.OrderHandler) *chi.Mux {
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
		r.Get("/instruments", edgeH.GetAvailableInstruments)
		r.Get("/snapshot/{token}/{date}", snapH.GetSnapshot)
		r.Get("/market-dna/{token}/{date}", edgeH.GetMarketDNA)

		// --- OMS Proxy Routes ---
		r.Post("/order/place", orderH.HandleProxy)
		r.Put("/order/modify", orderH.HandleProxy)
		r.Delete("/order/cancel", orderH.HandleProxy)
		r.Get("/order/state", orderH.HandleProxy)
	})
	return r
}
