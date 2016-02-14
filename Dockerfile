FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_LOCALFS_META_DATADIR /tmp/localfs
ENV CLAWIO_LOCALFS_META_TMPDIR /tmp/localfs
ENV CLAWIO_LOCALFS_META_PORT 57001
ENV CLAWIO_LOCALFS_META_LOGLEVEL "error"
ENV CLAWIO_LOCALFS_META_PROP "service-localfs-prop:57003"
ENV CLAWIO_LOCALFS_META_PROPMAXACTIVE 1024
ENV CLAWIO_LOCALFS_META_PROPMAXIDLE 1024
ENV CLAWIO_LOCALFS_META_PROPMAXCONCURRENCY 1024
ENV CLAWIO_SHAREDSECRET secret

ADD . /go/src/github.com/clawio/service-localfs-meta
WORKDIR /go/src/github.com/clawio/service-localfs-meta

RUN go get -u github.com/tools/godep
RUN godep restore
RUN go install

ENTRYPOINT /go/bin/service-localfs-meta

EXPOSE 57001

