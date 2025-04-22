# Use the official Golang image as the base image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the port your application runs on (default: 8080)
EXPOSE 8080

# Set environment variables (optional, can also be passed at runtime)
ENV PORT=8080

# Command to run the application
CMD ["./main"]