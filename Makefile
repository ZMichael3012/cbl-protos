BUF_VERSION ?= 1.47.2
DOCKER_BUF := docker run --rm -v "$(CURDIR)":/w -w /w bufbuild/buf:$(BUF_VERSION)

.PHONY: lint breaking generate build

lint:
	$(DOCKER_BUF) lint

# breaking сверяется с последней заморозкой (включая инкрементальные cf-N.M):
# перед выпуском новой версии заменить тег на актуальный.
breaking:
	$(DOCKER_BUF) breaking --against ".git#tag=cbl-protos@cf-3"

generate:
	go run github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION) generate

build:
	go build ./...
