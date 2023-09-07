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
# NOTE: if you use mount to add the certs symlinks that are not a sub directory will not work- this is a Dockerism
COPY cert/privkey1.pem    ./key.pem
COPY cert/fullchain1.pem  ./cert.pem


# Neither ENTRYPOINT nor CMD support ENV variables so manually keep EXPOSE and "-Port" in-sync
EXPOSE 443

ENTRYPOINT ["/zoomLogger", "-Port", "443"]
