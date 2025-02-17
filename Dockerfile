FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS build

WORKDIR /usr/src/app

ARG TARGETOS TARGETARCH
COPY . .

RUN ./scripts/install_helm.sh $TARGETARCH /usr/bin
RUN ./scripts/install_kustomize.sh $TARGETARCH /usr/bin

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
     GOOS=$TARGETOS GOARCH=$TARGETARCH make

FROM debian:bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    jq \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/src/app/yfg /usr/bin/yfg
COPY --from=build /usr/bin/helm /usr/bin/helm
COPY --from=build /usr/bin/kustomize /usr/bin/kustomize

ENTRYPOINT ["/usr/bin/yfg", "generate"]
