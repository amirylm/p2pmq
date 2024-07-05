# Build

FROM golang:1.20 AS builder

RUN apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -yq --no-install-recommends \
  git make g++ gcc-aarch64-linux-gnu wget \
  && rm -rf /var/lib/apt/lists/*

ARG APP_VERSION=nightly
ARG APP_NAME=pmq
ARG BUILD_TARGET=pmq

WORKDIR /pmq

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -tags netgo -a -v -o ./bin/${BUILD_TARGET} ./cmd/${BUILD_TARGET}

# Runtime

FROM alpine:latest as runner

ARG BUILD_TARGET=pmq

RUN apk --no-cache --upgrade add ca-certificates bash

WORKDIR /pmq

COPY --from=builder /pmq/.env* ./
COPY --from=builder /pmq/resources/config/*.pmq.yaml ./
COPY --from=builder /pmq/bin/${BUILD_TARGET} ./app

CMD ["./app"]