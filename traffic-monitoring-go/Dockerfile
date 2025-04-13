# Stage 1: Build the binary
FROM golang:1.19 AS builder

# Set the working directory to the module root
WORKDIR /workspace

# Copy go.mod and go.sum first, and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the entire project into the container
COPY . .

# Disable CGO to produce a static binary and build.
ENV CGO_ENABLED=0
# Build the binary from the module root
RUN go build -o traffic-monitoring-go ./app/main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Install CA certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root

# Copy the binary from the builder stage
COPY --from=builder /workspace/traffic-monitoring-go .

# Expose the port and run the binary
EXPOSE 8080
# DSRC PORT
EXPOSE 5001
# C-V2X PORT  
EXPOSE 5002
CMD ["./traffic-monitoring-go"]
