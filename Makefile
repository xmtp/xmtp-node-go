.PHONY: all build lint test coverage docker-image docker-image-multiarch

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
	go test ./...

# Set target-specific variables for docker images
docker-image docker-image-multiarch: DOCKER_IMAGE_TAG ?= latest
docker-image docker-image-multiarch: DOCKER_IMAGE_NAME ?= registry.digitalocean.com/xmtp-staging/xmtp-node-go:$(DOCKER_IMAGE_TAG)
docker-image docker-image-multiarch: GIT_COMMIT = $(shell git rev-parse HEAD)

# build a docker image
docker-image:
	docker build --tag $(DOCKER_IMAGE_NAME) \
		--build-arg="GIT_COMMIT=${GIT_COMMIT}" .

# Build for multiarch to support Apple Silicon.
# Assumes Docker Buildx is configured on the machine
# Also worth mentioning this is painfully slow due to QEMU
docker-image-multiarch:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag ${DOCKER_IMAGE_NAME} \
		--build-arg="GIT_COMMIT=${GIT_COMMIT}" \
		--push .
