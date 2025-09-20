# Multi-stage Dockerfile for cat-server
# Based on Alpine Linux for minimal size and security

# Stage 1: Build environment
FROM golang:alpine AS builder

# Set build directory
WORKDIR /build

# Install build dependencies (only git is needed for go modules)
RUN apk add --no-cache git

# Set environment variables for static linking
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Copy dependency files first (better caching)
COPY go.mod ./
# Copy go.sum if it exists (some projects may not have it)
COPY go.su[m] ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/ ./src/

# Build the application with optimization flags
RUN go build -ldflags="-w -s" -o cat-server src/main.go

# Verify the binary is statically linked
RUN ldd cat-server 2>&1 | grep -q "not a dynamic executable" || echo "WARNING: Binary may not be static"

# Stage 2: Runtime environment
FROM alpine:latest

# Install runtime dependencies
# wget is needed for health checks
RUN apk add --no-cache wget ca-certificates

# Create non-root user for security
# -D: Don't create home directory
# -H: Don't create home directory
# -s: Set shell to /sbin/nologin (no login allowed)
RUN adduser -D -H -s /sbin/nologin app

# Set working directory
WORKDIR /app

# Copy the binary from builder stage with proper ownership
COPY --from=builder --chown=app:app /build/cat-server ./cat-server

# Create files directory with proper permissions
RUN mkdir -p files && chown app:app files

# Switch to non-root user
USER app

# Expose port 8080
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set default command
CMD ["./cat-server"]