# Build stage
FROM golang:1.23.4-alpine AS builder
RUN apk add --no-cache git make build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN chmod 644 /app/SWIFT_CODES.csv
RUN go build -o main .

# Dockerfile
FROM golang:1.23.4-alpine AS tester

# Install dependencies
RUN apk add --no-cache git make build-base

WORKDIR /app
COPY . .
RUN go mod download
CMD ["go", "test", "-v", "-coverprofile=/app/coverage.txt", "-covermode=atomic", "./..."]


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/SWIFT_CODES.csv .

EXPOSE 8080
CMD ["/app/main"]
