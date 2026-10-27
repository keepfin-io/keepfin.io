
.PHONY: hot down build build-go build-web clean test auth-init

# Build everything (frontend + backend)
build:
	@echo "Building Vue frontend..."
	cd web && npm install && npm run build
	@echo "Building Go backend..."
	go build -o bin/keepnode .
	@echo "✅ Build completed successfully!"

# Build only Go backend
build-go:
	@echo "Building Go backend..."
	@mkdir -p bin
	go build -o bin/keepnode .
	@echo "✅ Go build completed!"

# Build only Vue frontend
build-web:
	@echo "Building Vue frontend..."
	cd web && npm install && npm run build
	@echo "✅ Vue build completed!"

# Run tests
test:
	@echo "Running Go tests..."
	go test ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/dist/
	@echo "✅ Clean completed!"

# Start development environment
auth-init:
	@sh scripts/init-auth-env.sh

hot: auth-init
	@echo "Starting development environment with Docker Compose (Vue + Go)..."
	@docker compose -p keepnode -f docker-compose.hot.yml up --build

# Stop development environment
down:
	@echo "Stopping development environment..."
	@docker compose -p keepnode -f docker-compose.hot.yml down

# Rebuild and restart
restart: down hot
