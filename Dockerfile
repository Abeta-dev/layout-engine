# syntax=docker/dockerfile:1.7

# -----------------------------------------------------------------------------
# Stage 1: Build static Go 1.26.0 binary (CGO_ENABLED=0; pure-Go SQLite + pgx)
# -----------------------------------------------------------------------------
FROM golang:1.26.0-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /out/layout-engine \
    ./cmd/layout-engine

# -----------------------------------------------------------------------------
# Stage 2: Minimal non-root runtime image
# -----------------------------------------------------------------------------
FROM alpine:3.21

LABEL org.opencontainers.image.title="abeta-layout-engine" \
      org.opencontainers.image.description="Abeta Dynamic SDUI Layout Engine & Visual Layout Studio" \
      org.opencontainers.image.source="https://github.com/Abeta-dev/layout-engine" \
      org.opencontainers.image.licenses="Apache-2.0"

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -g 10001 -S abeta \
    && adduser -u 10001 -S -G abeta -H -s /sbin/nologin abeta \
    && mkdir -p /data \
    && chown -R abeta:abeta /data

WORKDIR /app
COPY --from=builder /out/layout-engine /usr/local/bin/layout-engine

ENV SERVICE_NAME=abeta-layout-engine \
    PORT=8090 \
    GIN_MODE=release \
    DB_DIALECT=sqlite \
    DB_DSN="file:/data/layout-engine.db?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)" \
    AUTO_MIGRATE=true \
    SEED_STARTER_DATA=true \
    DEFAULT_TENANT=abeta-fresh

VOLUME ["/data"]
EXPOSE 8090

USER 10001:10001

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:${PORT}/healthz >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/usr/local/bin/layout-engine"]
