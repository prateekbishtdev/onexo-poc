.PHONY: run build test up down deps logs

-include .env
export

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

# Start only Postgres and Redis, for running the app on the host with `make run`.
deps:
	docker compose up -d postgres redis

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f app
