# build stage
FROM golang:1.25.4-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o back "./cmd"

# run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/back .
CMD ["./back"]
