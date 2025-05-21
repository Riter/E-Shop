# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o comments_service ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Установить curl
RUN apk add --no-cache curl

# Copy the binary from builder
COPY --from=builder /app/comments_service .
# Copy environment files
COPY environment/ ./environment/

# Expose the port
EXPOSE 30333

# Run the application
CMD ["./comments_service"] 