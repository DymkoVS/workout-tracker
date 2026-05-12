.PHONY: dev db run build tidy seed

dev:
	go run ./cmd/server

db:
	docker compose up db -d

run:
	docker compose up --build

build:
	go build -o ./bin/server ./cmd/server

tidy:
	go mod tidy

seed:
	go run ./cmd/seed

test:
	go test ./...
