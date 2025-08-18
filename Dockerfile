# syntax = docker/dockerfile:1.4

ARG TARGET=enduro
ARG GO_VERSION

FROM golang:${GO_VERSION}-alpine AS build-go
WORKDIR /src
ENV CGO_ENABLED=0
COPY --link go.* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY --link . .

FROM build-go AS build-enduro
ARG VERSION_PATH
ARG VERSION_LONG
ARG VERSION_SHORT
ARG VERSION_GIT_HASH
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go build \
	-trimpath \
	-ldflags="-X '${VERSION_PATH}.Long=${VERSION_LONG}' -X '${VERSION_PATH}.Short=${VERSION_SHORT}' -X '${VERSION_PATH}.GitCommit=${VERSION_GIT_HASH}'" \
	-o /out/enduro \
	./cmd/enduro

FROM build-go AS build-enduro-a3m-worker
ARG VERSION_PATH
ARG VERSION_LONG
ARG VERSION_SHORT
ARG VERSION_GIT_HASH
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go build \
	-trimpath \
	-ldflags="-X '${VERSION_PATH}.Long=${VERSION_LONG}' -X '${VERSION_PATH}.Short=${VERSION_SHORT}' -X '${VERSION_PATH}.GitCommit=${VERSION_GIT_HASH}'" \
	-o /out/enduro-a3m-worker \
	./cmd/enduro-a3m-worker

FROM build-go AS build-enduro-am-worker
ARG VERSION_PATH
ARG VERSION_LONG
ARG VERSION_SHORT
ARG VERSION_GIT_HASH
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go build \
	-trimpath \
	-ldflags="-X '${VERSION_PATH}.Long=${VERSION_LONG}' -X '${VERSION_PATH}.Short=${VERSION_SHORT}' -X '${VERSION_PATH}.GitCommit=${VERSION_GIT_HASH}'" \
	-o /out/enduro-am-worker \
	./cmd/enduro-am-worker

FROM debian:13-slim AS base
RUN apt-get update && apt-get install -y --no-install-recommends libxml2-utils
RUN groupadd --gid 1000 enduro \
	&& useradd --uid 1000 --gid 1000 -m enduro
USER enduro

FROM base AS enduro
COPY --link --from=build-enduro /out/enduro /home/enduro/bin/enduro
COPY --link --from=build-enduro /src/enduro.toml /home/enduro/.config/enduro.toml
COPY --link hack/xsd/premis.xsd /home/enduro/premis.xsd
RUN ["mkdir", "-m", "700", "-p", "/home/enduro/logs"]
CMD ["/home/enduro/bin/enduro", "--config", "/home/enduro/.config/enduro.toml"]

FROM base AS enduro-a3m-worker
COPY --link --from=build-enduro-a3m-worker /out/enduro-a3m-worker /home/enduro/bin/enduro-a3m-worker
COPY --link --from=build-enduro-a3m-worker /src/enduro.toml /home/enduro/.config/enduro.toml
COPY --link hack/xsd/premis.xsd /home/enduro/premis.xsd
CMD ["/home/enduro/bin/enduro-a3m-worker", "--config", "/home/enduro/.config/enduro.toml"]

FROM base AS enduro-am-worker
COPY --link --from=build-enduro-am-worker /out/enduro-am-worker /home/enduro/bin/enduro-am-worker
COPY --link --from=build-enduro-am-worker /src/enduro.toml /home/enduro/.config/enduro.toml
COPY --link hack/xsd/premis.xsd /home/enduro/premis.xsd
CMD ["/home/enduro/bin/enduro-am-worker", "--config", "/home/enduro/.config/enduro.toml"]

FROM ${TARGET}
