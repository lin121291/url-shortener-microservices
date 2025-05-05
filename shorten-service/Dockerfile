FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o shorten-service .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/shorten-service .

EXPOSE 8081

CMD ["./shorten-service"]