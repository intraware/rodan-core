FROM golang:1.24.1-alpine AS builder
RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o rodan .


FROM alpine:latest AS runner
RUN apk --no-cache add ca-certificates docker-cli

WORKDIR /app
COPY --from=builder /app/rodan .
COPY --from=builder /app/sample.config.toml ./config.toml

EXPOSE 8000
# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/ping || exit 1

CMD ["./rodan"]