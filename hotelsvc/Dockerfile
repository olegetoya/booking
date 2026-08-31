FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /out/hotelsvc ./cmd


FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /out/hotelsvc /app/hotelsvc
COPY config /app/config

EXPOSE 8080
EXPOSE 50051

ENTRYPOINT ["/app/hotelsvc"]
CMD ["-config=/app/config/docker.yaml"]