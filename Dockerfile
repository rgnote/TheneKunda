# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o thenekunday ./cmd/thenekunday

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/thenekunday .

# Create logs directory
RUN mkdir -p /app/logs

# Expose ports
EXPOSE 2222 2323 8080

# Run as non-root user
RUN adduser -D -u 1000 honeypot
USER honeypot

ENTRYPOINT ["./thenekunday"]
CMD []
