.PHONY: all build live backtest stop logs clean

APP_NAME = gidh-edge
BINARY = bin/$(APP_NAME)
MAIN_GO = cmd/server/main.go

# Default target
all: build

# Build the Go application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build -o $(BINARY) $(MAIN_GO)
	@echo "Build completed: $(BINARY)"

# Build and run PM2 in LIVE mode
live: build
	@echo "Starting PM2 in LIVE mode..."
	@pm2 describe $(APP_NAME)-live > /dev/null 2>&1 \
		&& MODE=live pm2 restart $(APP_NAME)-live --update-env \
		|| MODE=live pm2 start ./$(BINARY) --name $(APP_NAME)-live
	@pm2 save

# Build and run PM2 in BACKTEST mode
backtest: build
	@echo "Starting PM2 in BACKTEST mode..."
	@pm2 describe $(APP_NAME)-backtest > /dev/null 2>&1 \
		&& MODE=backtest pm2 restart $(APP_NAME)-backtest --update-env \
		|| MODE=backtest pm2 start ./$(BINARY) --name $(APP_NAME)-backtest
	@pm2 save

# Stop all PM2 instances for this app
stop:
	pm2 stop $(APP_NAME)-live || true
	pm2 stop $(APP_NAME)-backtest || true

# View PM2 logs
logs:
	pm2 logs

# Clean build artifacts
clean:
	@echo "Cleaning up build artifacts..."
	rm -rf bin/