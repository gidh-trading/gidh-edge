package router

import (
	"gidh-edge/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(edgeH *handler.EdgeHandler, snapH *handler.SnapshotHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Route("/api", func(r chi.Router) {
		r.Get("/instruments", edgeH.GetAvailableInstruments)
		r.Get("/baselines/{token}/{date}", edgeH.GetBaseline)
		r.Get("/snapshot/{token}/{date}", snapH.GetSnapshot)
	})
	return r
}
