FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

COPY shared/ ./shared/
COPY usersvc/ ./usersvc/

RUN cd usersvc && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o usersvc ./cmd/usersvc

# Use a minimal alpine image for the final image
FROM alpine:3.21

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/usersvc/usersvc .

# Expose the port
EXPOSE 8081

# Set the entry point
ENTRYPOINT ["./usersvc"]
