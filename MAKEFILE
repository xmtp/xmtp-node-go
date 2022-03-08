.PHONY: all build lint test coverage

all: build

deps: lint-install

build:
	go build -o build/xmtp main.go

vendor:
	go mod tidy

lint-install:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		bash -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint:
	@echo "lint"
	@golangci-lint --exclude=SA1019 run ./... --deadline=5m

test:
	go test

# build a docker image for the fleet
docker-image: DOCKER_IMAGE_TAG ?= latest
docker-image: DOCKER_IMAGE_NAME ?= registry.digitalocean.com/xmtp-staging/go-waku:$(DOCKER_IMAGE_TAG)
docker-image:
	docker build --tag $(DOCKER_IMAGE_NAME) \
		--build-arg="GIT_COMMIT=$(shell git rev-parse HEAD)" .