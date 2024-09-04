GO := go
GO_LINKER_FLAGS ?=
GO_BUILD_FLAGS ?=
IMAGE_TAG := main
DOCKER_FLAGS ?=
TEST_FLAGS ?=
TEST_PACKAGES ?= ./...
VERSION ?= $(shell git describe --tags --always)

all: yfg

.PHONY: yfg
yfg: schema
	$(GO) build -ldflags="-X 'github.com/chancez/yamlforge/cmd.Version=$(VERSION)' $(GO_LINKER_FLAGS)" $(GO_BUILD_FLAGS) -o yfg .

schema:
	$(GO) run ./tools/gen-jsonschema pkg/config/schema/schema.json

.PHONY: test
test:
	$(GO) test $(TEST_FLAGS) $(TEST_PACKAGES)

.PHONY: image
image:
	docker build $(DOCKER_FLAGS) -t quay.io/ecnahc515/yfg:$(IMAGE_TAG) .
