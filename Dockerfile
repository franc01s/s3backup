FROM golang:1.24-bullseye AS build
ENV DEBIAN_FRONTEND=noninteractive
ENV DEBIAN_FRONTEND=noninteractive
ARG ACTIONS_TOKEN

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends nmap upx unzip 
       
COPY . ./
RUN go mod download 

RUN go build -ldflags "-s -w" -o /server \
    && upx /server

FROM debian:12-slim
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates &&  /usr/sbin/update-ca-certificates && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY --from=build /server /server

ENTRYPOINT ["/server"]
