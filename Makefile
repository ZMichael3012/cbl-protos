BUF_VERSION ?= 1.47.2
DOCKER_BUF := docker run --rm -v "$(CURDIR)":/w -w /w bufbuild/buf:$(BUF_VERSION)

.PHONY: lint breaking generate build

lint:
	$(DOCKER_BUF) lint

# breaking compares against the last freeze (including the incremental cf-N.M):
# replace the tag with the current one before releasing a new version.
breaking:
	$(DOCKER_BUF) breaking --against ".git#tag=cbl-protos@cf-3.4"

# generate runs buf with the pinned version; the v prefix is required because
# the argument is resolved as a Go module version.
generate:
	go run github.com/bufbuild/buf/cmd/buf@v$(BUF_VERSION) generate

build:
	go build ./...
