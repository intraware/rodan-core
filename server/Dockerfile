FROM golang:1.24-alpine AS builder
RUN wget https://packages.timber.io/vector/0.37.0/vector-x86_64-unknown-linux-musl.tar.gz \
  && tar -xzf vector-x86_64-unknown-linux-musl.tar.gz \
  && mv vector-x86_64-unknown-linux-musl/bin/vector /usr/local/bin/
WORKDIR /app
COPY . .
RUN go build -o server

FROM alpine:latest AS runtime
COPY --from=builder /app/server /root/server
COPY --from=builder /usr/local/bin/vector /usr/local/bin/vector
CMD ["/root/server"]
