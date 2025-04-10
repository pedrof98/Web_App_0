FROM golang:1.22 AS builder

WORKDIR /app

# Copy go.mod and go.sum file to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . ./

# Build the application (pointing to the main.go location)
RUN CGO_ENABLED=0 GOOS=linux go build -o data-generator ./app

# Use a small image for the final stage
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/data-generator .

# Set default environment variables in a single layer
ENV SIEM_API_URL="http://app:8080" \
    EVENTS_PER_MINUTE="100" \
    ENABLE_ATTACK_SIMULATION="true" \
    ATTACK_FREQUENCY="30" \
    INCLUDE_V2X_EVENTS="true"

# Run the data generator
CMD ["./data-generator"]