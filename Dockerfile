FROM golang:1.18-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download  # Download dependencies

# Copy the rest of the code
COPY . .

# Build the Go application
RUN go build -o /snow cmd/main.go

FROM alpine:latest
COPY --from=builder /snow /snow
ENTRYPOINT ["/snow"]
