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
	db := postgres.MustConnect(cfg.DatabaseURL)
	defer db.Close()

	// 1. Initialize Layers
	repo := repo.NewPostgresRepo(db)
	engineClient := client.NewHTTPEngineClient(cfg.EngineURL)

	edgeSvc := service.NewEdgeService(repo)
	snapSvc := service.NewSnapshotService(repo, engineClient)

	edgeH := handler.NewEdgeHandler(edgeSvc)
	snapH := handler.NewSnapshotHandler(snapSvc)

	// 2. Start Server
	r := router.NewRouter(edgeH, snapH)
	logger.Infof("GIDH Edge Command Center listening on :%s", cfg.Port)
	http.ListenAndServe(":"+cfg.Port, r)
}
