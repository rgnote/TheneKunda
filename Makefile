.PHONY: all build run clean docker docker-build docker-run test fmt vet lint help

# Binary name
BINARY_NAME=thenekunday
DOCKER_IMAGE=thenekunday-honeypot

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

all: test build

## build: Build the application
build:
	@echo "Building..."
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/$(BINARY_NAME)

## run: Run the application
run: build
	@echo "Running..."
	./$(BINARY_NAME)

## clean: Clean build files
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf logs/

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker-compose up -d

## docker-stop: Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	docker-compose down

## docker-logs: View Docker logs
docker-logs:
	docker-compose logs -f

## generate-config: Generate example config
generate-config: build
	./$(BINARY_NAME) -generate-config

## help: Display this help message
help:
	@echo "TheneKunda Honeypot - Make commands:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
