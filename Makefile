.PHONY: build run test test-coverage lint docker-build docker-run docker-stop clean

BINARY  := pack-calculator
PORT    ?= 8080
IMAGE   := pack-calculator:latest

## build: compile the binary
build:
	go build -o $(BINARY) ./cmd/server

## run: compile and run locally
run: build
	PORT=$(PORT) ./$(BINARY)

## test: run all tests
test:
	go test ./...

## test-coverage: run tests and open an HTML coverage report
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

## lint: run go vet
lint:
	go vet ./...

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE) .

## docker-run: build and start the container (detached)
docker-run: docker-build
	docker run -d --name $(BINARY) -p $(PORT):8080 -e PORT=8080 $(IMAGE)
	@echo "Running at http://localhost:$(PORT)"

## docker-stop: stop and remove the container
docker-stop:
	docker stop $(BINARY) && docker rm $(BINARY)

## compose-up: start with docker-compose
compose-up:
	docker compose up --build -d

## compose-down: stop docker-compose stack
compose-down:
	docker compose down

## clean: remove build artefacts
clean:
	rm -f $(BINARY) coverage.out coverage.html
