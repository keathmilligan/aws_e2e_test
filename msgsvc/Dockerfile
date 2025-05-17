FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o msgsvc ./cmd/msgsvc

# Use a small alpine image
FROM alpine:3.21

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/msgsvc .

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./msgsvc"]
