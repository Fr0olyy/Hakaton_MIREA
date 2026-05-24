.PHONY: tidy fmt build run migrate-up docker-up docker-down agent-build agent-run

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

build:
	go build ./...

run:
	go run ./cmd/api

agent-build:
	cd services/agent-service && go build ./...

agent-run:
	cd services/agent-service && go run ./cmd/server

migrate-up:
	go run ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
