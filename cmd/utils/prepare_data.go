package main

import (
	"context"
	"gidh-edge/internal/repo"
	"gidh-edge/pkg/config"
	"gidh-edge/pkg/logger"
	"gidh-edge/pkg/postgres"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.App.LogLevel)

	db, err := postgres.New(cfg.DB.ConnString)
	if err != nil {
		logger.Errorf("Failed to connect to DB: %v", err)
		return
	}
	defer db.Close()

	r := repo.NewPostgresRepo(db)

	// 1. Get current valid dates from DB (those that have DNA)
	dnaDates, err := r.GetDNADates(context.Background())
	if err != nil {
		logger.Errorf("Failed to fetch DNA dates: %v", err)
		return
	}

	// 2. Scan the backup directory for archives
	files, err := os.ReadDir(cfg.App.BacktestBackupDir)
	if err != nil {
		logger.Errorf("Failed to read backup directory: %v", err)
		return
	}

	for _, file := range files {
		name := file.Name()
		if !strings.HasPrefix(name, "backup_") || !strings.HasSuffix(name, ".tar.xz") {
			continue
		}

		dateStr := strings.TrimSuffix(strings.TrimPrefix(name, "backup_"), ".tar.xz")

		// CLEANUP: If backup exists but date is NOT in DB, delete both backup and extracted data
		if !dnaDates[dateStr] {
			logger.Warnf("Sync: Date %s no longer in DB. Cleaning up files...", dateStr)
			os.Remove(filepath.Join(cfg.App.BacktestBackupDir, name))
			os.RemoveAll(filepath.Join(cfg.App.BacktestDataDir, dateStr))
			continue
		}

		// PREPARATION: If date is in DB, ensure it is extracted to BACKTEST_DATA_DIR
		targetDir := filepath.Join(cfg.App.BacktestDataDir, dateStr)
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			logger.Infof("Extracting data for %s to %s...", dateStr, cfg.App.BacktestDataDir)

			tarPath := filepath.Join(cfg.App.BacktestBackupDir, name)
			// Extracting directly to the data directory
			cmd := exec.Command("tar", "-xJf", tarPath, "-C", cfg.App.BacktestDataDir)

			if output, err := cmd.CombinedOutput(); err != nil {
				logger.Errorf("Failed to untar %s: %v. Output: %s", name, err, string(output))
			} else {
				logger.Infof("Successfully prepared %s", dateStr)
			}
		}
	}
	logger.Info("Maintenance and preparation complete.")
}
