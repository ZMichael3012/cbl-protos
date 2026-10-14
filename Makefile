BUF_VERSION ?= 1.47.2
DOCKER_BUF := docker run --rm -v "$(CURDIR)":/w -w /w bufbuild/buf:$(BUF_VERSION)

.PHONY: lint generate build

lint:
	$(DOCKER_BUF) lint

generate:
	go run github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION) generate

build:
	go build ./...
