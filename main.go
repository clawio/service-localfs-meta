package main

import (
	"fmt"
	"github.com/clawio/grpcxlog"
	pb "github.com/clawio/service.localstore.meta/proto"
	"github.com/rs/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
	"os"
	"strconv"
)

var log xlog.Logger

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
	log.Infof("%s=%s", dataDirEnvar, e.dataDir)
	log.Infof("%s=%s", tmpDirEnvar, e.tmpDir)
	log.Infof("%s=%d", portEnvar, e.port)
	log.Infof("%s=%s", sharedSecretEnvar, "******")
}

func setupLog() {

	// Install the logger handler with default output on the console
	lh := xlog.NewHandler(xlog.LevelDebug)

	// Set some global env fields
	host, _ := os.Hostname()
	lh.SetFields(xlog.F{
		"svc":  serviceID,
		"host": host,
	})

	// Plug the xlog handler's input to Go's default logger
	grpclog.SetLogger(grpcxlog.Log{lh.NewLogger()})

	log = lh.NewLogger()
}

func main() {

	setupLog()

	log.Infof("Service %s started", serviceID)

	env, err := getEnviron()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	printEnviron(env)

	p := &newServerParams{}
	p.dataDir = env.dataDir
	p.tmpDir = env.tmpDir
	p.sharedSecret = env.sharedSecret

	srv := newServer(p)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", env.port))
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLocalServer(grpcServer, srv)
	grpcServer.Serve(lis)
}
