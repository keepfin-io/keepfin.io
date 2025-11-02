
.PHONY: hot down build
build:
	@echo "Building Svelte frontend..."
	cd web && npm install && npm run build
	@echo "Building Go backend..."
	go build -o bin/keepnode .

hot:
	@echo "Starting development environment with Docker Compose..."
	@echo "Running Docker Compose..."
	@docker compose -p keepnode -f docker-compose.hot.yml up --build

down:
	@echo "Stopping development environment..."
	@docker compose -p keepnode -f docker-compose.hot.yml down
