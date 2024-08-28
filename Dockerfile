# syntax = docker/dockerfile:1.4

ARG TARGET=enduro
ARG GO_VERSION

FROM alpine:3.20 AS build-libxml
RUN apk add --no-cache libxml2-utils

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

FROM gcr.io/distroless/base-debian12:latest AS base
USER 1000

FROM base AS enduro
COPY --from=build-enduro --link /out/enduro /home/enduro/bin/enduro
COPY --from=build-enduro --link /src/enduro.toml /home/enduro/.config/enduro.toml
CMD ["/home/enduro/bin/enduro", "--config", "/home/enduro/.config/enduro.toml"]

FROM base AS enduro-a3m-worker
COPY --from=build-enduro-a3m-worker --link /out/enduro-a3m-worker /home/enduro/bin/enduro-a3m-worker
COPY --from=build-enduro-a3m-worker --link /src/enduro.toml /home/enduro/.config/enduro.toml
CMD ["/home/enduro/bin/enduro-a3m-worker", "--config", "/home/enduro/.config/enduro.toml"]
COPY hack/xsd/premis.xsd /home/enduro/premis.xsd
COPY --from=build-libxml /usr/bin/xmllint /usr/bin/xmllint
COPY --from=build-libxml /usr/lib/libxml2.so.2 /usr/lib/libxml2.so.2
COPY --from=build-libxml /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY --from=build-libxml /lib/libz.so.1 /lib/libz.so.1
COPY --from=build-libxml /usr/lib/liblzma.so.5 /usr/lib/liblzma.so.5

FROM base AS enduro-am-worker
COPY --from=build-enduro-am-worker --link /out/enduro-am-worker /home/enduro/bin/enduro-am-worker
COPY --from=build-enduro-am-worker --link /src/enduro.toml /home/enduro/.config/enduro.toml
CMD ["/home/enduro/bin/enduro-am-worker", "--config", "/home/enduro/.config/enduro.toml"]
COPY hack/xsd/premis.xsd /home/enduro/premis.xsd
COPY --from=build-libxml /usr/bin/xmllint /usr/bin/xmllint
COPY --from=build-libxml /usr/lib/libxml2.so.2 /usr/lib/libxml2.so.2
COPY --from=build-libxml /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY --from=build-libxml /lib/libz.so.1 /lib/libz.so.1
COPY --from=build-libxml /usr/lib/liblzma.so.5 /usr/lib/liblzma.so.5

FROM ${TARGET}
