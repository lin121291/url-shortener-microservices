FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o redirect-service .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/redirect-service .

EXPOSE 8082

CMD ["./redirect-service"]