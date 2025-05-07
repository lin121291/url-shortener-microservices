FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shorten-service/go.mod shorten-service/go.sum ./
COPY common ../common
RUN go mod download

COPY shorten-service .
RUN CGO_ENABLED=0 GOOS=linux go build -o shorten-service .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/shorten-service .

EXPOSE 8081

CMD ["./shorten-service"]