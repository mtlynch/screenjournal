FROM golang:1.26.1 AS builder

ARG TARGETPLATFORM

COPY ./announce /app/announce
COPY ./auth /app/auth
COPY ./build /app/build
COPY ./cmd /app/cmd
COPY ./dev-scripts/build-backend /app/dev-scripts/build-backend
COPY ./email /app/email
COPY ./handlers /app/handlers
COPY ./markdown /app/markdown
COPY ./metadata /app/metadata
COPY ./passwordreset /app/passwordreset
COPY ./random /app/random
COPY ./ratelimit /app/ratelimit
COPY ./screenjournal /app/screenjournal
COPY ./store /app/store
COPY ./.git /app/.git
COPY ./go.* /app/

WORKDIR /app

RUN TARGETPLATFORM="${TARGETPLATFORM}" \
      ./dev-scripts/build-backend prod

FROM scratch AS artifact
COPY --from=builder /app/bin/screenjournal ./

FROM litestream/litestream:0.3.13 AS litestream

FROM alpine:3.15

RUN apk add --no-cache bash tzdata

ARG TZ
RUN if [[ -n "${TZ}" ]]; then \
      ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime && \
      echo "${TZ}" > /etc/timezone; \
    fi

COPY --from=builder /app/bin/screenjournal /app/screenjournal
COPY --from=litestream /usr/local/bin/litestream /app/litestream
COPY ./docker-entrypoint /app/docker-entrypoint
COPY ./litestream.yml /etc/litestream.yml
COPY ./LICENSE /app/LICENSE

WORKDIR /app

ENV DB_PATH=/data/store.db

ENTRYPOINT ["/app/docker-entrypoint"]
