
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY api-gateway/go.mod api-gateway/go.sum ./
RUN go mod download

COPY api-gateway .
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]