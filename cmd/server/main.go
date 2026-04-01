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

	db, err := postgres.New(cfg.DB.ConnString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 1. Initialize Layers
	repoInstance := repo.NewPostgresRepo(db)
	engineClient := client.NewHTTPEngineClient(cfg.API.EngineURL)

	edgeSvc := service.NewEdgeService(repoInstance, engineClient)
	snapSvc := service.NewSnapshotService(repoInstance, engineClient)

	edgeH := handler.NewEdgeHandler(edgeSvc)
	snapH := handler.NewSnapshotHandler(snapSvc)

	// 2. Start Server
	r := router.NewRouter(edgeH, snapH)
	logger.Infof("GIDH Edge Command Center listening on :%s", cfg.API.Port)
	http.ListenAndServe(":"+cfg.API.Port, r)
}
