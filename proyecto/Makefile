DATABASE_URL ?= postgres://appuser:apppass@localhost:5432/myapp?sslmode=disable

## —— Docker ——————————————————————————————————————

.PHONY: up
up: ## Levanta todos los contenedores (db + api)
	docker compose up --build -d

.PHONY: down
down: ## Baja todos los contenedores
	docker compose down

.PHONY: down-clean
down-clean: ## Baja contenedores y elimina volúmenes (reset de datos)
	docker compose down -v

.PHONY: logs
logs: ## Muestra logs de todos los servicios
	docker compose logs -f

.PHONY: logs-api
logs-api: ## Muestra logs solo de la API
	docker compose logs -f api

## —— Base de datos ————————————————————————————————

.PHONY: db-up
db-up: ## Levanta solo la base de datos
	docker compose up db -d --wait

.PHONY: db-shell
db-shell: ## Abre una sesión psql en la base de datos
	docker compose exec db psql -U appuser -d myapp

.PHONY: db-migrate
db-migrate: db-up ## Ejecuta migraciones manualmente (ya corren al iniciar el contenedor)
	docker compose exec db psql -U appuser -d myapp -f /docker-entrypoint-initdb.d/001_create_users.sql

## —— Desarrollo local ————————————————————————————

.PHONY: dev
dev: db-up ## Levanta la DB y corre la API localmente con hot-reload
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/api

.PHONY: build
build: ## Compila el binario
	go build -o bin/api ./cmd/api

.PHONY: test
test: ## Ejecuta los tests
	go test ./...

.PHONY: test-integration
test-integration: db-up ## Ejecuta tests de integración contra la DB real
	DATABASE_URL=$(DATABASE_URL) go test -tags=integration ./...

.PHONY: lint
lint: ## Ejecuta el linter
	golangci-lint run ./...

## —— Utilidades ——————————————————————————————————

.PHONY: curl-users
curl-users: ## Prueba el endpoint GET /users
	curl -s http://localhost:8080/users | jq .

.PHONY: help
help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
