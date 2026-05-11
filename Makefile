# Hephaestus — top-level Makefile
# Convenient one-liners for the most common operations

.PHONY: help up down restart logs demo test-chaincode lint clean api dashboard

PROJECT_NAME := hephaestus
COMPOSE_FILE := docker/docker-compose.yaml

help: ## Show this help
	@echo "Hephaestus — Permissioned Blockchain for Bank Reporting"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

up: ## Bring the full Fabric network online (5 banks + BI + OJK)
	@echo "🔨 Forging Hephaestus network..."
	@bash network/scripts/generate-artifacts.sh
	@docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up -d
	@echo "⏳ Waiting for peers to be ready..."
	@sleep 10
	@bash network/scripts/create-channel.sh
	@bash network/scripts/deploy-chaincode.sh
	@echo "✅ Network is up. Dashboard: http://localhost:8080  API: http://localhost:3000"

down: ## Tear down the network and clean state
	@docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) down -v
	@rm -rf network/crypto-config network/channel-artifacts network/system-genesis-block
	@echo "🧹 Network down and state cleared."

restart: down up ## down + up

logs: ## Tail logs from all peers and orderers
	@docker compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) logs -f --tail=50

demo: ## Submit a sample LBU report and validate it
	@bash scripts/demo.sh

test-chaincode: ## Run chaincode unit tests
	@cd chaincode/antasena && go test -v -cover ./...

api: ## Run API gateway only (assumes network is up)
	@cd api-gateway && npm install && npm start

dashboard: ## Open the dashboard in default browser
	@xdg-open http://localhost:8080 2>/dev/null || open http://localhost:8080

lint: ## Lint all source code
	@cd chaincode/antasena && gofmt -l . && go vet ./...
	@cd api-gateway && npx prettier --check src/
	@ansible-lint ansible/ || true

clean: down ## down + remove built artifacts
	@docker volume prune -f
	@rm -rf api-gateway/node_modules api-gateway/wallet/*
