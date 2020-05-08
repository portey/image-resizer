export GO111MODULE=on
export GOSUMDB=off

IMAGE_TAG := $(shell git rev-parse HEAD)
DOCKER_REPO = nexus.tools.devopenocean.studio
SHELL=/bin/bash

.PHONY: ci
ci: lint test_unit test_integration build

.PHONY: deps
deps:
	go mod download
	go mod vendor

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -o artifacts/svc .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test_unit
test_unit:
	go test -cover -v `go list ./...`

.PHONY: test_integration
test_integration:
	INTEGRATION_TEST=YES go test -cover -v `go list ./...`

.PHONY: dockerise
dockerise:
	docker build -t "${DOCKER_REPO}/image-resizer:${IMAGE_TAG}" .