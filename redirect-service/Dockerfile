FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY redirect-service/go.mod redirect-service/go.sum ./
COPY common ../common
RUN go mod download

COPY redirect-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o redirect-service .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/redirect-service .

EXPOSE 8082

CMD ["./redirect-service"]