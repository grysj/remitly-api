FROM golang:1.23.4-alpine AS builder
RUN apk add --no-cache git make build-base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM golang:1.23.4-alpine AS tester
RUN apk add --no-cache git make build-base
WORKDIR /app
COPY . .
RUN go mod download


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/SWIFT_CODES.csv .
EXPOSE 8080
CMD ["/app/main"]
