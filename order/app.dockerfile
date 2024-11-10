# Use a multi-stage build to keep the final image slim
FROM golang:latest AS build

# Set the working directory to /app
WORKDIR /app

# Copy the go.mod and go.sum files from the root to /app
COPY go.mod go.sum ./

# Download the module dependencies
RUN go mod download

# Copy the entire project, as Go needs access to the other services' code in the root structure
COPY . .

# Build the order service binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./order/cmd/order

# Use a minimal base image for the final image
FROM alpine:latest
WORKDIR /root/
COPY --from=build /app/main .

EXPOSE 8080
CMD ["./main"]
