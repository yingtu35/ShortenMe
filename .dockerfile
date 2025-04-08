# Use the official Go image
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS base

# Set the working directory in the container
WORKDIR /app

# Copy the go.mod and go.sum files to the container
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Install golangci-lint
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2

# Copy rest of the source code to the container
COPY . .

FROM base AS lint

# Run linter
RUN golangci-lint run --timeout=5m

FROM base AS builder

# Build the application with additional flags for better security and performance
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o shortenMe cmd/app/main.go

# Use a smaller image for the runtime
FROM --platform=$TARGETPLATFORM alpine:3.19 AS runner

# Create a non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/shortenMe .

# Copy template files from the builder stage
COPY --from=builder /app/templates ./templates

# Set proper permissions
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the port
EXPOSE 8080

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/health || exit 1

# Specify the command to run the application
CMD ["./shortenMe"]
