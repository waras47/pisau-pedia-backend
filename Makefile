include .env
export

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

seed:
	go run ./cmd/seed

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)" up

migrate-down:
	migrate -path migrations -database "mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)" down 1

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

lint:
	golangci-lint run

test:
	go test ./... -cover

.PHONY: run build seed migrate-create migrate-up migrate-down docker-up docker-down docker-logs lint test
