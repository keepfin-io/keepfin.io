
.PHONY: hot down build build-go build-web clean test auth-init search-init

DEV_COMPOSE := docker compose -p keepnode -f docker-compose.hot.yml

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

search-init:
	@sh scripts/init-search-env.sh

hot: auth-init search-init
	@echo "Starting development environment with Docker Compose (Vue + Go)..."
	@$(DEV_COMPOSE) config --quiet
	@$(DEV_COMPOSE) build keepnode auth
	@set -eu; \
		$(DEV_COMPOSE) stop vue keepnode auth; \
		setup_complete=false; \
		trap 'if [ "$$setup_complete" = false ]; then $(DEV_COMPOSE) stop vue keepnode auth >/dev/null 2>&1 || true; fi' EXIT; \
		trap 'exit 130' INT; \
		trap 'exit 143' TERM; \
		$(DEV_COMPOSE) up --detach --wait --wait-timeout 60 --remove-orphans postgres; \
		$(DEV_COMPOSE) exec -T \
			-e PGUSER=postgres -e PGDATABASE=keepnode -e PGPASSWORD=postgres \
			-e KEEPFIN_DB_PASSWORD=keepfin-dev-runtime \
			-e KEEPFIN_MIGRATION_DB_PASSWORD=keepfin-dev-migrator \
			-e AUTH_DB_PASSWORD=auth-dev-runtime \
			-e AUTH_MIGRATION_DB_PASSWORD=auth-dev-migrator \
			postgres sh /security/roles.sh; \
		$(DEV_COMPOSE) up --detach --no-deps --force-recreate auth keepnode; \
		$(DEV_COMPOSE) exec -T \
			-e DB_USERNAME=keepfin_migrator -e DB_PASSWORD=keepfin-dev-migrator \
			keepnode go run . migrate; \
		$(DEV_COMPOSE) exec -T \
			-e DATABASE_URL=postgres://auth_migrator@postgres:5432/keepnode \
			-e PGPASSWORD=auth-dev-migrator -e AUTH_MAIL_MODE=disabled \
			auth npm run migrate:dev; \
		$(DEV_COMPOSE) exec -T \
			-e PGUSER=postgres -e PGDATABASE=keepnode -e PGPASSWORD=postgres \
			postgres psql -X --no-password -f /security/permissions.sql; \
		$(DEV_COMPOSE) exec -T \
			-e DATABASE_URL=postgres://postgres@postgres:5432/keepnode \
			-e PGPASSWORD=postgres -e AUTH_MAIL_MODE=disabled \
			auth node src/seed.ts; \
		$(DEV_COMPOSE) exec -T auth touch /tmp/keepfin-ready; \
		$(DEV_COMPOSE) exec -T keepnode touch /tmp/keepfin-ready; \
		setup_complete=true; \
		exec $(DEV_COMPOSE) up --no-build postgres typesense auth keepnode vue

# Stop development environment
down:
	@echo "Stopping development environment..."
	@$(DEV_COMPOSE) down

# Rebuild and restart
restart: down hot
