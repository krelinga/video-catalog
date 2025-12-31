# Build stage
FROM golang:1.25 AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o video-catalog-server ./server

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/video-catalog-server .

# Expose port (adjust if needed)
EXPOSE 8080

# Run the binary
CMD ["./video-catalog-server"]
