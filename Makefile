# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=clickety-clack
OUTPUT_DIR=bin
GO_VERSION=$(shell grep -E '^go [0-9]+\.[0-9]+' go.mod | awk '{print $$2}')

# Docker Compose parameters
DOCKER_COMPOSE=docker compose
DOCKER_COMPOSE_FILE=compose.yaml

# Run
.PHONY: run
run:
	go run . --soundsDir=./sounds/YUNZII_C68_-_AnDr3W --config=./sounds/YUNZII_C68_-_AnDr3W/config.json

# Build targets
.PHONY: all
all: clean build

.PHONY: build
build: build-linux build-windows build-macos

.PHONY: build-linux
build-linux:
	GO_VERSION=$(GO_VERSION) BINARY_NAME=$(BINARY_NAME) $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) run --rm --build build-linux

.PHONY: build-windows
build-windows:
	GO_VERSION=$(GO_VERSION) BINARY_NAME=$(BINARY_NAME) $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) run --rm --build build-windows

.PHONY: build-macos
build-macos:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(OUTPUT_DIR)/$(BINARY_NAME)_darwin_amd64

.PHONY: clean
clean:
	$(GOCLEAN)
	$(GOCMD) mod tidy
	rm -rf $(OUTPUT_DIR)
	GO_VERSION=$(GO_VERSION) BINARY_NAME=$(BINARY_NAME) $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) rm -f --stop --volumes
