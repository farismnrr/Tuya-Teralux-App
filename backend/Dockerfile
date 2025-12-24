# Start with a lightweight Go base image
# Using Debian-based image to avoid QEMU/Alpine compatibility issues (execve error)
# and to support Go 1.24 required by dependencies.
FROM golang:1.24-bookworm AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache for dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
# -o main: name the output binary "main"
RUN go build -o main .

# Download and install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Start a new stage for a smaller final image
# Using debian-slim is more robust for multi-arch builds on QEMU than Alpine
FROM debian:bookworm-slim

# Install ca-certificates and mysql-client for migrations
RUN apt-get update && apt-get install -y ca-certificates default-mysql-client && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy golang-migrate binary
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy migration files
COPY --from=builder /app/domain/teralux/migrations ./migrations

# Copy entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Set default auto-migrate value (can be overridden at build time)
ARG AUTO_MIGRATE=false
ENV AUTO_MIGRATE=${AUTO_MIGRATE}

# Expose the port the app runs on
EXPOSE 8080

# Use entrypoint script to run migrations before starting the app
ENTRYPOINT ["./entrypoint.sh"]
