GO := go
GO_LINKER_FLAGS ?=
GO_BUILD_FLAGS ?=
IMAGE_TAG := main
DOCKER_FLAGS ?=

all: yfg

.PHONY: yfg
yfg:
	$(GO) build --ldflags='$(GO_LINKER_FLAGS)' $(GO_BUILD_FLAGS) -o yfg .

.PHONY: image
image:
	docker build $(DOCKER_FLAGS) -t quay.io/ecnahc515/yfg:$(IMAGE_TAG) .
