FROM golang:1.25-trixie AS build
ENV DEBIAN_FRONTEND=noninteractive
ARG ACTIONS_TOKEN

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends nmap upx-ucl unzip 
       
COPY . ./
RUN go mod download 

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o /server .
 #   && upx /server

FROM gcr.io/distroless/static-debian13:nonroot
#FROM debian:13-slim
ENV DEBIAN_FRONTEND=noninteractive

# Install CA certificates for HTTPS/TLS connections
#RUN apt-get update && apt-get install -y --no-install-recommends \
#    ca-certificates \
#    && update-ca-certificates \
#    && apt-get clean \
#    && rm -rf /var/lib/apt/lists/*
WORKDIR /

COPY --from=build /server /server

ENTRYPOINT ["/server"]
