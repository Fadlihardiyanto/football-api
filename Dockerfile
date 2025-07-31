FROM golang:1.23-bullseye AS builder

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    librdkafka-dev \
    git \
    pkg-config \
    libc6-dev \
    make \
    ca-certificates

# Set workdir
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the web binary
WORKDIR /app/cmd/web
RUN go build -o /app/bin/web .

# Build the worker binary
WORKDIR /app/cmd/worker
RUN go build -o /app/bin/worker .

# ---------- RUNTIME STAGE ----------
FROM debian:bullseye-slim

# Install only runtime dependencies
RUN apt-get update && apt-get install -y librdkafka1 ca-certificates && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Set timezone
ENV TZ=Asia/Jakarta

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/web /web
COPY --from=builder /app/bin/worker /worker

# Ensure executables
RUN chmod +x /web /worker

# Copy optional resource dirs if needed
COPY --from=builder /app/uploads /app/uploads

# Set default CMD (can be overridden in docker-compose)
CMD ["/web"]
