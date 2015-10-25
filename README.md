# service.localstore.meta
Microservice responsible for local storage metadata

It contains:

* a gRPC server
* a gRPC client

To run the server define the following enviromental variables or `source environ`

```
export CLAWIO_LOCALSTORE_DATADIR=/tmp
export CLAWIO_LOCALSTORE_TMPDIR=/tmp
export CLAWIO_LOCALSTORE_PORT=57001
```
