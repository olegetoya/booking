FROM golang:1.25.5-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -o /out/bookingsvc ./cmd


FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /out/bookingsvc /app/bookingsvc
COPY config /app/config

EXPOSE 8081

ENTRYPOINT ["/app/bookingsvc"]
CMD ["-config=/app/config/docker.yaml"]