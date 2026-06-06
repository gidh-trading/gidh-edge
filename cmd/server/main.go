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

	repoInstance := repo.NewPostgresRepo(db)
	orderRepoInstance := repo.NewOrderRepository(db)
	engineClient := client.NewHTTPEngineClient(cfg.API.EngineURL)

	// Initialize lean services
	edgeSvc := service.NewEdgeService(repoInstance, engineClient)
	snapSvc := service.NewSnapshotService(repoInstance, engineClient)
	backtestSvc := service.NewBacktestService(engineClient, repoInstance, cfg.App.BacktestBackupDir)
	orderSvc := service.NewOrderService(engineClient, orderRepoInstance, cfg.Kite.APIKey, cfg.Kite.AccessToken)

	// Initialize handlers
	edgeH := handler.NewEdgeHandler(edgeSvc)
	snapH := handler.NewSnapshotHandler(snapSvc)
	backtestH := handler.NewBacktestHandler(backtestSvc)
	orderH := handler.NewOrderHandler(orderSvc, edgeSvc, cfg.App.Mode)

	// Pass all three handlers to the router
	r := router.NewRouter(edgeH, snapH, backtestH, orderH)

	logger.Infof("GIDH Edge Command Center listening on :%s", cfg.API.Port)
	http.ListenAndServe(":"+cfg.API.Port, r)
}
