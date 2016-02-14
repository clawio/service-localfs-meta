package main

import (
	"fmt"
	pb "github.com/clawio/service-localfs-meta/proto/metadata"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"runtime"
	"strconv"
)

const (
	serviceID               = "CLAWIO_LOCALFS_META"
	dataDirEnvar            = serviceID + "_DATADIR"
	tmpDirEnvar             = serviceID + "_TMPDIR"
	portEnvar               = serviceID + "_PORT"
	propEnvar               = serviceID + "_PROP"
	logLevelEnvar           = serviceID + "_LOGLEVEL"
	propMaxActiveEnvar      = serviceID + "_PROPMAXACTIVE"
	propMaxIdleEnvar        = serviceID + "_PROPMAXIDLE"
	propMaxConcurrencyEnvar = serviceID + "_PROPMAXCONCURRENCY"
	sharedSecretEnvar       = "CLAWIO_SHAREDSECRET"
)

type environ struct {
	dataDir            string
	tmpDir             string
	port               int
	prop               string
	logLevel           string
	propMaxActive      int
	propMaxIdle        int
	propMaxConcurrency int
	sharedSecret       string
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

	e.prop = os.Getenv(propEnvar)
	e.logLevel = os.Getenv(logLevelEnvar)

	propMaxActive, err := strconv.Atoi(os.Getenv(propMaxActiveEnvar))
	if err != nil {
		return nil, err
	}
	e.propMaxActive = propMaxActive

	propMaxIdle, err := strconv.Atoi(os.Getenv(propMaxIdleEnvar))
	if err != nil {
		return nil, err
	}
	e.propMaxIdle = propMaxIdle

	propMaxConcurrency, err := strconv.Atoi(os.Getenv(propMaxConcurrencyEnvar))
	if err != nil {
		return nil, err
	}
	e.propMaxConcurrency = propMaxConcurrency

	e.sharedSecret = os.Getenv(sharedSecretEnvar)
	return e, nil
}
func printEnviron(e *environ) {
	log.Infof("%s=%s\n", dataDirEnvar, e.dataDir)
	log.Infof("%s=%s\n", tmpDirEnvar, e.tmpDir)
	log.Infof("%s=%d\n", portEnvar, e.port)
	log.Infof("%s=%s\n", propEnvar, e.prop)
	log.Infof("%s=%d\n", propMaxActiveEnvar, e.propMaxActive)
	log.Infof("%s=%d\n", propMaxIdleEnvar, e.propMaxIdle)
	log.Infof("%s=%d\n", propMaxConcurrencyEnvar, e.propMaxConcurrency)
	log.Infof("%s=%s\n", sharedSecretEnvar, "******")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	env, err := getEnviron()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	printEnviron(env)

	p := &newServerParams{}
	p.dataDir = env.dataDir
	p.tmpDir = env.tmpDir
	p.prop = env.prop
	p.sharedSecret = env.sharedSecret
	p.propMaxActive = env.propMaxActive
	p.propMaxIdle = env.propMaxIdle
	p.propMaxConcurrency = env.propMaxConcurrency

	l, err := log.ParseLevel(env.logLevel)
	if err != nil {
		l = log.ErrorLevel
	}
	log.SetLevel(l)

	log.Infof("Service %s started", serviceID)
	printEnviron(env)

	// Create data and tmp dirs
	if err := os.MkdirAll(p.dataDir, 0644); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	if err := os.MkdirAll(p.tmpDir, 0644); err != nil {
		log.Error(err)
		os.Exit(1)
	}

	srv := newServer(p)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", env.port))
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMetaServer(grpcServer, srv)
	grpcServer.Serve(lis)
}
