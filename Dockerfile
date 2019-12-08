FROM golang:alpine AS build

ENV DISTRIBUTION_DIR /go/src/github.com/rolinux/hs110-exporter

ARG GOOS=linux
ARG GOARCH=amd64

RUN set -ex \
    && apk add --no-cache git

WORKDIR $DISTRIBUTION_DIR
COPY . $DISTRIBUTION_DIR
RUN go get -v ./...
RUN CGO_ENABLED=0 go build -v  -o hs110-exporter hs110-exporter.go

FROM alpine

COPY --from=build /go/src/github.com/rolinux/hs110-exporter/hs110-exporter /bin/hs110-exporter
EXPOSE 9498
CMD ["hs110-exporter"]
