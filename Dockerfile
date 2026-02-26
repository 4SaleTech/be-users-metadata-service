# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service ./cmd

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g "" appuser
USER appuser

WORKDIR /app
COPY --from=builder /app/service .

ENTRYPOINT ["./service"]
