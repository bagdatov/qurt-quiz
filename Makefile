BINARY     := qurt-quiz
BUILD_DIR  := bin
FRONTEND   := frontend
STATIC_DIR := $(FRONTEND)/dist

.PHONY: all build build-backend build-frontend install dev test lint clean docker-build docker-up

all: build

## install — install frontend npm dependencies
install:
	cd $(FRONTEND) && npm install

## build — compile backend binary and frontend bundle
build: build-backend build-frontend

build-backend:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/server

build-frontend:
	cd $(FRONTEND) && npm run build

## dev — run backend and frontend concurrently (requires 'make install' first)
dev:
	@which concurrently > /dev/null 2>&1 || npm install -g concurrently
	concurrently \
		--names "backend,frontend" \
		--prefix-colors "cyan,magenta" \
		"go run ./cmd/server" \
		"cd $(FRONTEND) && npm run dev"

## run — run the compiled backend serving the built frontend
run: build
	./$(BUILD_DIR)/$(BINARY) -static $(STATIC_DIR)

## test — run all Go tests
test:
	go test ./...

## test-verbose — run Go tests with verbose output
test-verbose:
	go test -v ./...

## lint — run Go vet and frontend ESLint
lint:
	go vet ./...
	cd $(FRONTEND) && npm run lint

## vet — alias for Go vet only
vet:
	go vet ./...

## tidy — tidy Go modules
tidy:
	go mod tidy

## clean — remove build artefacts
clean:
	rm -rf $(BUILD_DIR) $(STATIC_DIR)

## docker-build — build the production Docker image
docker-build:
	docker build -t $(BINARY):latest .

## docker-up — start full stack with Docker Compose
docker-up:
	docker compose up --build

## docker-down — stop Docker Compose services
docker-down:
	docker compose down

## help — list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## //'
