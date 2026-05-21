.PHONY: tidy fmt build run migrate-up docker-up docker-down

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal

build:
	go build ./...

run:
	go run ./cmd/api

migrate-up:
	go run ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v
