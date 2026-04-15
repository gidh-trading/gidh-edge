package service

import (
	"context"
	"gidh-edge/internal/client"
	"io"
	"net/http"
)

type BacktestService struct {
	engine *client.HTTPEngineClient
}

func NewBacktestService(e *client.HTTPEngineClient) *BacktestService {
	return &BacktestService{engine: e}
}

func (s *BacktestService) ProxyBacktestRequest(ctx context.Context, method, uri string, body io.Reader, headers http.Header) (*http.Response, error) {
	return s.engine.ForwardRawRequest(ctx, method, uri, body, headers)
}
