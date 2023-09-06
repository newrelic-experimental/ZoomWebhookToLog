# syntax=docker/dockerfile:1

# Step 1: Build the application from source
FROM golang:1.21-alpine AS build-stage

WORKDIR /build

# The Go image does not include git, add it to Alpine
RUN apk add git

RUN git clone https://github.com/newrelic-experimental/ZoomWebhookToLog.git

WORKDIR ZoomWebhookToLog

# Install the application's Go dependencies
RUN go mod download

# Build the executable
RUN GOARCH=amd64 GOOS=linux go build -o /zoomLogger internal/main.go

# Step 2: Deploy the application binary into a lean image
FROM alpine AS build-release-stage

WORKDIR /

COPY --from=build-stage /zoomLogger /zoomLogger

# COPY certs in from a local location
COPY scratch/cert.pem .
COPY scratch/key.pem .

# Neither ENTRYPOINT nor CMD support ENV variables so manually keep EXPOSE and "-Port" in-sync
EXPOSE 443

ENTRYPOINT ["/zoomLogger", "-Port", "443"]