# Base image for building purposes
FROM golang:1.14-alpine as builder

RUN apk --no-cache add alsa-lib-dev git alpine-sdk

WORKDIR /jellycli

ARG JELLYCLI_BRANCH=master

# use caching layers as needed

RUN git clone -b ${JELLYCLI_BRANCH} --single-branch --depth 1 https://github.com/tryffel/jellycli ./

RUN go mod download

RUN go build . && ./jellycli -help


# Alpine runtime
FROM alpine:3.10

RUN apk --no-cache add alsa-lib-dev dbus-x11 alsa-utils
COPY --from=builder /jellycli/jellycli /usr/local/bin

RUN mkdir /root/.config/
ENTRYPOINT [ "jellycli" ]