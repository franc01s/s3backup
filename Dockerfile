FROM golang:1.24-bullseye AS build
ENV DEBIAN_FRONTEND=noninteractive
ENV DEBIAN_FRONTEND=noninteractive
ARG ACTIONS_TOKEN

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends nmap upx unzip
       
COPY . ./
RUN go mod download 

RUN GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o /server \
    && upx /server

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /server /server

ENTRYPOINT ["/server"]
