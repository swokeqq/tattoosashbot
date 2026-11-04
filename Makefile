-include .env
export

.PHONY: run test lint

run:
	go run ./cmd/bot

test:
	go test -race ./...

lint:
	golangci-lint run