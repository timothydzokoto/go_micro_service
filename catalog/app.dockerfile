# Stage 1: Build the Go application
FROM golang:latest AS build

# Set the working directory in the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire catalog service directory into the container
COPY ./catalog /app/catalog

# Build the application for Linux (CGO_ENABLED=0 for cross-compilation)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./catalog/cmd/catalog

# Stage 2: Create the runtime image
FROM alpine:latest

# Install libc6-compat to support Go binaries
RUN apk add --no-cache libc6-compat

# Set the working directory in the runtime container
WORKDIR /root/

# Copy the compiled binary from the build stage
COPY --from=build /app/main .

# Expose the port your application will run on
EXPOSE 8080

# Start the application
CMD ["./main"]
