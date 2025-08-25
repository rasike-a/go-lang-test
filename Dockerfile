# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install ca-certificates (needed for HTTPS requests)
RUN apk --no-cache add ca-certificates git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Set environment variable for port
ENV PORT=8080

# Run the binary
CMD ["./main"]