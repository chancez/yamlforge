GO := go
GO_LINKER_FLAGS ?=
GO_BUILD_FLAGS ?=
IMAGE_TAG := main
DOCKER_FLAGS ?=
TEST_FLAGS ?=
TEST_PACKAGES ?= ./...
VERSION ?= $(shell git describe --tags --always)
BINARY ?= yfg

default: yfg

all: check test yfg

.PHONY: yfg
yfg: schema yfg-bin

.PHONY: yfg-bin
yfg-bin:
	$(GO) build -ldflags="-X 'github.com/chancez/yamlforge/cmd.Version=$(VERSION)' $(GO_LINKER_FLAGS)" $(GO_BUILD_FLAGS) -o $(BINARY) .

schema:
	$(GO) run ./tools/gen-jsonschema pkg/config/schema/schema.json

.PHONY: test
test:
	$(GO) test $(TEST_FLAGS) $(TEST_PACKAGES)

.PHONY: check
check:
	golangci-lint run

.PHONY: image
image:
	docker build $(DOCKER_FLAGS) -t quay.io/ecnahc515/yfg:$(IMAGE_TAG) .

.PHONY: release
release:
	for GOOS in darwin linux windows; do \
		for GOARCH in arm64 amd64; do \
			env GOOS=$$GOOS GOARCH=$$GOARCH $(MAKE) yfg-bin BINARY=yfg-$$GOOS-$$GOARCH; \
		done \
	done

