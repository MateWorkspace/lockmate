# syntax=docker/dockerfile:1

FROM golang:1.26.5-alpine AS builder

WORKDIR /src/lockmate

# Compose supplies the sibling module as a named build context.
COPY --from=mate-golang go.mod go.sum /src/mate-golang/
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY --from=mate-golang pgxdt /src/mate-golang/pgxdt
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1

COPY database/seeder ./database/seeder
COPY cmd ./cmd
COPY internal ./internal
COPY docs/swagger ./docs/swagger

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -o /out/lockmate ./cmd/main

FROM alpine:3.23 AS runtime

RUN apk add --no-cache ca-certificates jq \
    && addgroup -S -g 10001 lockmate \
    && adduser -S -u 10001 -G lockmate lockmate

WORKDIR /app

COPY --from=builder /out/lockmate /app/lockmate
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY database/migrations/*.up.sql database/migrations/*.down.sql /app/migrations/
COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

ARG LOCKMATE_PORT_HTTP_SERVER=10701
ARG LOCKMATE_PORT_GRPC_SERVER=10702

ENV LOCKMATE_PORT_HTTP_SERVER=${LOCKMATE_PORT_HTTP_SERVER} \
    LOCKMATE_PORT_GRPC_SERVER=${LOCKMATE_PORT_GRPC_SERVER}

EXPOSE ${LOCKMATE_PORT_HTTP_SERVER} ${LOCKMATE_PORT_GRPC_SERVER}

USER lockmate

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${LOCKMATE_PORT_HTTP_SERVER}/api/version" || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
