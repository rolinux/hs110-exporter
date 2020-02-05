FROM golang AS build

ENV DISTRIBUTION_DIR /go/src/github.com/rolinux/hs110-exporter

RUN apt-get update && apt-get install -y --no-install-recommends \
		git \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR $DISTRIBUTION_DIR
COPY . $DISTRIBUTION_DIR
RUN go get -v ./...
RUN CGO_ENABLED=0 go build -v -a -installsuffix cgo -o hs110-exporter hs110-exporter.go

# run container with app on top on scratch empty container
FROM scratch

COPY --from=build /go/src/github.com/rolinux/hs110-exporter/hs110-exporter /bin/hs110-exporter

EXPOSE 9498

CMD ["hs110-exporter"]
