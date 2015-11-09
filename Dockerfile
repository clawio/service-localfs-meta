FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_LOCALSTOREMETA_DATADIR /tmp
ENV CLAWIO_LOCALSTOREMETA_TMPDIR /tmp
ENV CLAWIO_LOCALSTOREMETA_PORT 57001
ENV CLAWIO_LOCALSTOREMETA_PROP "service.localstore.prop"
ENV CLAWIO_SHAREDSECRET secret

RUN go get -u github.com/clawio/service.localstore.meta

ENTRYPOINT /go/bin/service.localstore.meta

EXPOSE 57001

