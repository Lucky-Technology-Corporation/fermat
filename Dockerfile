# First stage: Build the Go application
FROM golang:alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Mod and Sum files
COPY go.mod go.sum ./

# Download necessary dependencies and tools
RUN apk add --no-cache git
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Second stage: Create the final minimal image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose port 3000 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
