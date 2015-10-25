# service.localstore.meta
Microservice responsible for local storage metadata

It contains:

* a gRPC server
* a gRPC client

## Server

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

## Client

The following snippet is used to create a folder.

````
package main

import (
	"github.com/clawio/service.localstore.meta/lib"
	pb "github.com/clawio/service.localstore.meta/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func main() {
	p := &lib.NewClientParams{}
	p.Addr = "localhost:57001"
	p.Opts = []grpc.DialOption{grpc.WithInsecure()}

	client, err := lib.NewClient(p)
	if err != nil {
		log.Fatal(err)
	}

	idt := &pb.Identity{}
	idt.Pid = "hugo"
	idt.Idp = "localhost"
	idt.DisplayName = "Hugo González Labrador"

	mkdirReq := &pb.MkdirReq{}
	mkdirReq.Idt = idt
	mkdirReq.Path = "somefolder"

	ctx := context.Background()

	_, err = client.Mkdir(ctx, mkdirReq)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Folder %s created", mkdirReq.Path)
}
```

Check the documentation and find more examples inside the client tests.



