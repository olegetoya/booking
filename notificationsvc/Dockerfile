
# Stage 1: Build
FROM golang:1.26.0-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -o /out/notificationsvc ./cmd


FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/notificationsvc /app/notificationsvc
COPY config /app/config

ENTRYPOINT ["/app/notificationsvc"]
CMD ["-config=/app/config/docker.yaml"]
