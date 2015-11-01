package main

import (
	"fmt"
	pb "github.com/clawio/service.localstore.meta/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
)

const (
	serviceID         = "CLAWIO_LOCALSTOREMETA"
	dataDirEnvar      = serviceID + "_DATADIR"
	tmpDirEnvar       = serviceID + "_TMPDIR"
	portEnvar         = serviceID + "_PORT"
	sharedSecretEnvar = "CLAWIO_SHAREDSECRET"
)

type environ struct {
	dataDir      string
	tmpDir       string
	port         int
	sharedSecret string
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
	e.sharedSecret = os.Getenv(sharedSecretEnvar)
	return e, nil
}
func printEnviron(e *environ) {
	log.Printf("%s=%s", dataDirEnvar, e.dataDir)
	log.Printf("%s=%s", tmpDirEnvar, e.tmpDir)
	log.Printf("%s=%d", portEnvar, e.port)
	log.Printf("%s=%s", sharedSecretEnvar, "******")
}
func main() {
	log.Printf("Service %s started", serviceID)

	env, err := getEnviron()
	if err != nil {
		log.Fatal(err)
	}
	
	printEnviron(env)

	p := &newServerParams{}
	p.dataDir = env.dataDir
	p.tmpDir = env.tmpDir
	p.sharedSecret = env.sharedSecret

	srv := newServer(p)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", env.port))
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLocalServer(grpcServer, srv)
	grpcServer.Serve(lis)
}
