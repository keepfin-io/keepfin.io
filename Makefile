
.PHONY: hot down build

build:
	@echo "Building Vue frontend..."
	cd web && npm install && npm run build
	@echo "Building Go backend..."
	go build -o bin/keepnode .

hot:
	@echo "Starting development environment with Docker Compose (Vue + Go)..."
	@docker compose -p keepnode -f docker-compose.hot.yml up --build

down:
	@echo "Stopping development environment..."
	@docker compose -p keepnode -f docker-compose.hot.yml down
