FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_LOCALSTOREMETA_DATADIR /tmp
ENV CLAWIO_LOCALSTOREMETA_TMPDIR /tmp
ENV CLAWIO_LOCALSTOREMETA_PORT 57001
ENV CLAWIO_LOCALSTOREMETA_PROP "service-localstore-prop"
ENV CLAWIO_SHAREDSECRET secret

ADD . /go/src/clawio/service.localstore.meta

RUN go get -u github.com/tools/godep
RUN cd /go/src/clawio/service.localstore.meta
RUN godep restore
RUN go install

ENTRYPOINT /go/bin/service.localstore.meta

EXPOSE 57001

