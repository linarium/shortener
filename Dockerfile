# Стадия сборки
FROM golang:1.23.8-alpine3.20 AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /shortener ./cmd/shortener

FROM alpine:3.20

COPY --from=builder /shortener /shortener

ENTRYPOINT ["/shortener"]