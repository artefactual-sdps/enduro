# syntax = docker/dockerfile:1.4

ARG TARGET=enduro

FROM golang:1.18.4-alpine AS build-go
WORKDIR /src
ENV CGO_ENABLED=0
COPY --link go.* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY --link . .

FROM build-go AS build-enduro
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go build -o /out/enduro .

FROM build-go AS build-enduro-a3m-worker
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go build -o /out/enduro-a3m-worker ./cmd/enduro-a3m-worker

FROM alpine:3.16.0 AS base
ARG USER_ID=1000
ARG GROUP_ID=1000
RUN addgroup -g ${GROUP_ID} -S enduro
RUN adduser -u ${USER_ID} -S -D enduro enduro
USER enduro

FROM base AS enduro
COPY --from=build-enduro --link /out/enduro /home/enduro/bin/enduro
COPY --from=build-enduro --link /src/enduro.toml /home/enduro/.config/enduro.toml
CMD ["/home/enduro/bin/enduro", "--config", "/home/enduro/.config/enduro.toml"]

FROM base AS enduro-a3m-worker
COPY --from=build-enduro-a3m-worker --link /out/enduro-a3m-worker /home/enduro/bin/enduro-a3m-worker
COPY --from=build-enduro-a3m-worker --link /src/enduro.toml /home/enduro/.config/enduro.toml
CMD ["/home/enduro/bin/enduro-a3m-worker", "--config", "/home/enduro/.config/enduro.toml"]

FROM ${TARGET}
