FROM golang:1.7-alpine

RUN apk upgrade --update
RUN apk add git
RUN mkdir -p /usr/local/go/src/github.com/alaa/catalog/
Add . /usr/local/go/src/github.com/alaa/catalog/

WORKDIR /usr/local/go/src/github.com/alaa/catalog/
RUN go get -v
RUN go build .
ENTRYPOINT catalog

EXPOSE 8080
