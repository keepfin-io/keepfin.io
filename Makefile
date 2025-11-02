.PHONY: hot

hot:
	@echo "Starting Go and Svelte servers..."
	@cd web && npm run dev & \
	air
