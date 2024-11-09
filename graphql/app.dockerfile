FROM golang:latest AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main ./graphql

FROM alpine:latest
RUN apk add --no-cache libc6-compat
WORKDIR /root/
COPY --from=build /app/main .
EXPOSE 8080
CMD ["./main"]
