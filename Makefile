# Go parameters
GOCMD=/usr/local/go/bin/go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=server
BINARY_PATH=bin/$(BINARY_NAME)

.PHONY: all build clean test deps run

all: deps build

build:
	$(GOBUILD) -o $(BINARY_PATH) cmd/server/main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)

test:
	$(GOTEST) -v ./...

deps:
	$(GOMOD) tidy
	$(GOMOD) download

run: build
	./$(BINARY_PATH)

dev:
	$(GOCMD) run cmd/server/main.go

docker-build:
	docker build -t contextkeeper-backend .

docker-run:
	docker-compose up --build

docker-down:
	docker-compose down

.DEFAULT_GOAL := build