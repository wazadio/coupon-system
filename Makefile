.PHONY: help build up down restart logs test test-scenarios clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Docker containers
	docker-compose build

up: ## Start the application (with build)
	docker-compose up --build

up-detached: ## Start the application in detached mode
	docker-compose up -d --build

down: ## Stop and remove containers
	docker-compose down

restart: ## Restart the application
	docker-compose restart

logs: ## View application logs
	docker-compose logs -f

logs-api: ## View API logs only
	docker-compose logs -f api

logs-db: ## View database logs only
	docker-compose logs -f postgres

test: up-detached wait-for-api test-scenarios ## Run all tests (starts system if not running)

test-unit: ## Run unit tests with coverage
	@echo "Running unit tests..."
	go test -v -cover -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage: ## Run tests and show coverage summary
	@echo "Running tests with coverage..."
	go test -cover ./internal/...

test-scenarios: ## Run the critical scenario tests
	@echo "Running Flash Sale and Double Dip tests..."
	go test -v -timeout 60s ./test/...

test-flash-sale: ## Run only the Flash Sale test
	@echo "Running Flash Sale Attack test..."
	go test -v -timeout 30s ./test -run TestFlashSaleScenario

test-double-dip: ## Run only the Double Dip test
	@echo "Running Double Dip Attack test..."
	go test -v -timeout 30s ./test -run TestDoubleDipScenario

wait-for-api: ## Wait for API to be ready
	@echo "Waiting for API to be ready..."
	@for i in {1..30}; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "API is ready!"; \
			sleep 2; \
			exit 0; \
		fi; \
		echo "Waiting... ($$i/30)"; \
		sleep 2; \
	done; \
	echo "API failed to start in time"; \
	exit 1

health: ## Check API health status
	@curl -s http://localhost:8080/health | jq . || echo "API is not responding"

clean: down ## Clean up everything (containers, volumes, images)
	docker-compose down -v
	docker system prune -f

clean-all: ## Clean up everything including images
	docker-compose down -v --rmi all
	docker system prune -af

dev: ## Start in development mode with logs
	docker-compose up --build

status: ## Show container status
	docker-compose ps

shell-api: ## Open shell in API container
	docker exec -it coupon_api sh

shell-db: ## Open PostgreSQL shell
	docker exec -it coupon_db psql -U coupon_user -d coupon_db

# Example API calls
example-create: ## Example: Create a test coupon
	@echo "Creating test coupon..."
	@curl -X POST http://localhost:8080/api/coupons \
		-H "Content-Type: application/json" \
		-d '{"name":"TEST_COUPON","amount":10}' | jq .

example-claim: ## Example: Claim a coupon
	@echo "Claiming coupon..."
	@curl -X POST http://localhost:8080/api/coupons/claim \
		-H "Content-Type: application/json" \
		-d '{"user_id":"test_user","coupon_name":"TEST_COUPON"}' | jq .

example-get: ## Example: Get coupon details
	@echo "Getting coupon details..."
	@curl http://localhost:8080/api/coupons/TEST_COUPON | jq .
