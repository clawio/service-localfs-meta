# service.localstore.meta
Microservice responsible for local storage metadata

It contains:

* a gRPC server
* a gRPC client

To install the server do

```
go get -u github.com/clawio/service.localstore.meta
```

Then, define the following enviromental variables accordingly to your needs

```
export CLAWIO_LOCALSTORE_DATADIR=/tmp
export CLAWIO_LOCALSTORE_TMPDIR=/tmp
export CLAWIO_LOCALSTORE_PORT=57001
```

Run it 
```
$ service.localstore.meta
2015/10/25 22:52:49 Service CLAWIO_LOCALSTORE started
2015/10/25 22:52:49 CLAWIO_LOCALSTORE_DATADIR=/tmp
2015/10/25 22:52:49 CLAWIO_LOCALSTORE_TMPDIR=/tmp
2015/10/25 22:52:49 CLAWIO_LOCALSTORE_PORT=57001
```

To use the client


