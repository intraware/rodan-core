FROM golang:1.24.6-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o rodan .

FROM gcr.io/distroless/cc AS runner
WORKDIR /root
COPY --from=builder /app/rodan .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/root/rodan"]
