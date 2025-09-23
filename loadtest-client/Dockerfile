# Multi-stage build for load test client
FROM golang:1.23-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o loadtest \
    cmd/loadtest/main.go

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 loadtest && \
    adduser -D -s /bin/sh -u 1001 -G loadtest loadtest

WORKDIR /home/loadtest

# Copy binary and configs
COPY --from=builder /app/loadtest .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/README.md .

# Change ownership
RUN chown -R loadtest:loadtest /home/loadtest

# Switch to non-root user
USER loadtest

# Create results directory
RUN mkdir -p results

# Default command
CMD ["./loadtest", "--config", "configs/default.yaml"]
