FROM golang:1.5
MAINTAINER Hugo González Labrador


ADD . /go/src/github.com/service.localstore.meta
RUN go get -u github.com/clawio/service.localstore.meta
RUN . /go/src/github.com/service.localstore.meta/environ

ENTRYPOINT /go/bin/service.localstore.meta

EXPOSE 57001

