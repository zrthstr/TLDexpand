# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Run tests
RUN go test -v

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o tldexpand .

# Final stage - minimal image
FROM alpine:latest

# Install CA certificates for HTTPS (needed for DNS)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/tldexpand .

# Copy TLD list files
COPY --from=builder /build/*.txt .

# Set the binary as entrypoint
ENTRYPOINT ["/app/tldexpand"]

# Default: show help
CMD ["-h"]
