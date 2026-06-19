.PHONY: build test run docker-build docker-up docker-down fmt vet tidy

build:
	go build -o bin/whook ./cmd/whook

test:
	go test ./...

run:
	go run ./cmd/whook

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

docker-build:
	docker build -t whook:latest .

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
