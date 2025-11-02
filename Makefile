.PHONY: hot down

hot:
	@echo "Starting development environment with Docker Compose..."
	@echo "Running Docker Compose..."
	@docker compose -p keepnode -f docker-compose.hot.yml up --build

down:
	@echo "Stopping development environment..."
	@docker compose -p keepnode -f docker-compose.hot.yml down
