FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY shared/ ./shared/
COPY msgsvc/ ./msgsvc/

RUN cd msgsvc && CGO_ENABLED=0 GOOS=linux go build -mod=mod -a -installsuffix cgo -o msgsvc ./cmd/msgsvc

# Use a small alpine image
FROM alpine:3.21

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/msgsvc/msgsvc .

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./msgsvc"]
