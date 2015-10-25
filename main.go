package main

import (
	"fmt"
	"github.com/clawio/service.localstore.meta/Godeps/_workspace/src/google.golang.org/grpc"
	pb "github.com/clawio/service.localstore.meta/proto"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	serviceID    = "CLAWIO_LOCALSTORE"
	dataDirEnvar = serviceID + "_DATADIR"
	tmpDirEnvar  = serviceID + "_TMPDIR"
	portEnvar    = serviceID + "_PORT"
)

type environ struct {
	dataDir string
	tmpDir  string
	port    int
}

func getEnviron() (*environ, error) {
	e := &environ{}
	e.dataDir = os.Getenv(dataDirEnvar)
	e.tmpDir = os.Getenv(tmpDirEnvar)
	port, err := strconv.Atoi(os.Getenv(portEnvar))
	if err != nil {
		return nil, err
	}
	e.port = port
	return e, nil
}
func printEnviron(e *environ) {
	log.Printf("%s=%s", dataDirEnvar, e.dataDir)
	log.Printf("%s=%s", tmpDirEnvar, e.tmpDir)
	log.Printf("%s=%d", portEnvar, e.port)
}
func main() {
	log.Printf("Service %s started", serviceID)

	env, err := getEnviron()
	printEnviron(env)

	if err != nil {
		log.Fatal(err)
	}

	p := &newServerParams{}
	p.dataDir = env.dataDir
	p.tmpDir = env.tmpDir

	srv := newServer(p)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", env.port))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLocalServer(grpcServer, srv)
	grpcServer.Serve(lis)
}
