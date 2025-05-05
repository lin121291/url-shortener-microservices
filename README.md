# URL Shortener Microservices

A microservices-based URL shortening system written in Go, designed to explore scalable architecture and production-level service composition.

---

## Architecture Overview

```
                          +----------------+
                          |     Client     |
                          +--------+-------+
                                   |
                                   v
                         +---------+----------+
                         |      API Gateway   |        (Port: 8080)
                         +----+-----------+---+
                              |           |
              POST /shorten  |           | GET /:code
                              v           v
               +--------------+        +----------------+
               | Shorten Service |      | Redirect Service |
               |   (Port: 8081)  |      |   (Port: 8082)    |
               +--------+-------+      +---------+--------+
                        |                        |
              +---------v--------+     +--------v---------+
              |     PostgreSQL    |     |      Redis       |
              +------------------+     +------------------+

                          ⬇️ (Future Extension)
                    +------------------------+
                    |   Analytics Service    |
                    +------------------------+
```

---

##  Components

### API Gateway
- Entry point for all clients
- Routes: `/shorten` and `/:code`
- Forwards requests to the appropriate backend service

### Shorten Service
- Accepts long URL and returns a short code
- Stores mapping in PostgreSQL and cache in Redis

### Redirect Service
- Resolves the short code to the original long URL
- Handles redirects and optionally logs traffic to analytics

### Redis
- Used for fast cache lookups of URL mappings

### PostgreSQL
- Stores persistent short ↔ long URL mappings

### Analytics Service (Optional)
- Tracks visits, IP address, and metadata
- Could be event-driven with Kafka or RabbitMQ

---

## Tech Stack
- Golang (Gin framework)
- Redis (caching)
- PostgreSQL (data persistence)
- Docker / Docker Compos
- k8s