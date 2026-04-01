package service

import (
	"context"
	"gidh-edge/internal/client"
	"gidh-edge/internal/repo"
)

type EdgeService struct {
	repo   *repo.PostgresRepo
	engine client.EngineClient
}

func NewEdgeService(r *repo.PostgresRepo, e client.EngineClient) *EdgeService {
	return &EdgeService{repo: r, engine: e}
}

// GetEngineStatus determines if the stream is active or in post-mortem
func (s *EdgeService) GetEngineStatus(ctx context.Context) string {
	_, err := s.engine.GetActiveState(ctx, 0) // Ping with dummy token
	if err != nil {
		return "post-mortem"
	}
	return "active"
}
