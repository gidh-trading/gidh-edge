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

	// 2. CORS Middleware
	r.Use(cors.Handler(cors.Options{
		// AllowOriginFunc allows your Cloud UI to connect from any origin
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/instruments", edgeH.GetAvailableInstruments)
		r.Get("/baselines/{token}/{date}", edgeH.GetBaseline)
		r.Get("/snapshot/{token}/{date}", snapH.GetSnapshot)
	})
	return r
}
