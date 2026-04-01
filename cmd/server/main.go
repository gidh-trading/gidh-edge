package main

import (
	"gidh-edge/internal/client"
	"gidh-edge/internal/handler"
	"gidh-edge/internal/repo"
	"gidh-edge/internal/router"
	"gidh-edge/internal/service"
	"gidh-edge/pkg/config"
	"gidh-edge/pkg/logger"
	"gidh-edge/pkg/postgres"
	"net/http"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.App.LogLevel)

	// 1. Core Infrastructure
	db, _ := postgres.New(cfg.DB.ConnString)
	engineClient := client.NewHTTPEngineClient("http://localhost:8081")
	repository := repo.NewPostgresRepo(db)

	// 2. Services
	edgeSvc := service.NewEdgeService(repository, engineClient)
	snapSvc := service.NewSnapshotService(repository, engineClient)

	// 3. Handlers
	edgeHdl := handler.NewEdgeHandler(edgeSvc)
	snapHdl := handler.NewSnapshotHandler(snapSvc)

	// 4. Router
	r := router.NewRouter(edgeHdl, snapHdl)

	logger.Infof("Edge Command Center starting on port %s", cfg.API.Port)
	http.ListenAndServe(":"+cfg.API.Port, r)
}
