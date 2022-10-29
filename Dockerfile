FROM golang:1.19.1 AS builder

ARG TARGETPLATFORM

COPY ./cmd /app/cmd
COPY ./handlers /app/handlers
COPY ./store /app/store
COPY ./*.go /app/
COPY ./go.* /app/

WORKDIR /app

RUN set -x && \
    if [ "$TARGETPLATFORM" = "linux/arm/v7" ]; then \
      GOARCH="arm"; \
    elif [ "$TARGETPLATFORM" = "linux/arm64" ]; then \
      GOARCH="arm64"; \
    else \
      GOARCH="amd64"; \
    fi && \
    set -u && \
    GOOS=linux \
    go build \
      -tags netgo \
      -ldflags '-w -extldflags "-static"' \
      -o /app/screenjournal \
      cmd/screenjournal/main.go

FROM alpine:3.15

RUN apk add --no-cache bash

COPY --from=builder /app/screenjournal /app/screenjournal
COPY ./docker-entrypoint /app/docker-entrypoint
COPY ./static /app/static
COPY ./LICENSE /app/LICENSE

WORKDIR /app

ENTRYPOINT ["/app/docker-entrypoint"]
