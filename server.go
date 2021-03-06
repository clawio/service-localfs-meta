package main

import (
	"fmt"
	authlib "github.com/clawio/service-auth/lib"
	pb "github.com/clawio/service-localfs-meta/proto/metadata"
	proppb "github.com/clawio/service-localfs-meta/proto/propagator"
	"github.com/dropbox/godropbox/resource_pool"
	rus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"mime"
	"os"
	"path"
	"strings"
	"time"
)

const (
	dirPerm = 0755
)

var (
	unauthenticatedError = grpc.Errorf(codes.Unauthenticated, "identity not found")
	permissionDenied     = grpc.Errorf(codes.PermissionDenied, "access denied")
)

type newServerParams struct {
	dataDir            string
	tmpDir             string
	prop               string
	propMaxActive      int
	propMaxIdle        int
	propMaxConcurrency int
	sharedSecret       string
}

func newServer(p *newServerParams) *server {
	poolOptions := resource_pool.Options{}
	poolOptions.MaxActiveHandles = int32(p.propMaxActive)
	poolOptions.MaxIdleHandles = uint32(p.propMaxIdle)
	poolOptions.OpenMaxConcurrency = p.propMaxConcurrency
	poolOptions.Open = func(resourceLocation string) (interface{}, error) {
		con, err := grpc.Dial(resourceLocation, grpc.WithInsecure())
		if err != nil {
			rus.Error(err)
			return nil, err
		}
		return con, nil
	}
	poolOptions.Close = func(handle interface{}) error {
		con, ok := handle.(*grpc.ClientConn)
		if !ok {
			err := fmt.Errorf("connection handle is %+v but expected %+v", handle, "*grpc.ClientConn")
			rus.Error(err)
			return err
		}
		err := con.Close()
		if err != nil {
			rus.Error(err)
			return err
		}
		return nil
	}
	pool := resource_pool.NewSimpleResourcePool(poolOptions)
	pool.Register(p.prop)
	s := &server{}
	s.p = p
	s.grpcPool = pool
	return s
}

type server struct {
	p        *newServerParams
	grpcPool resource_pool.ResourcePool
}

func (s *server) Home(ctx context.Context, req *pb.HomeReq) (*pb.Void, error) {

	traceID, err := getTraceID(ctx)
	if err != nil {
		rus.Error(err)
		return &pb.Void{}, err
	}
	log := rus.WithField("trace", traceID).WithField("svc", serviceID)
	ctx = newTraceContext(ctx, traceID)

	log.Info("request started")

	// Time request
	reqStart := time.Now()

	defer func() {
		// Compute request duration
		reqDur := time.Since(reqStart)

		// Log access info
		log.WithFields(rus.Fields{
			"method":   "home",
			"type":     "grpcaccess",
			"duration": reqDur.Seconds(),
		}).Info("request finished")

	}()

	idt, err := authlib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	log.Infof("%s", idt)

	home := getHome(idt)

	log.Infof("user home is %s", home)

	pp := s.getPhysicalPath(home)

	log.Infof("user physical home is %s", pp)

	resource, err := s.grpcPool.Get("")
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	defer resource.Release()

	handle, err := resource.Handle()
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}
	con := handle.(*grpc.ClientConn)
	log.Infof("created connection to %s", s.p.prop)

	client := proppb.NewPropClient(con)

	_, err = os.Stat(pp)

	// Create home dir if not exists
	if os.IsNotExist(err) {

		log.Infof("user physical home %s does not exist", pp)

		err = os.MkdirAll(pp, dirPerm)
		if err != nil {
			log.Error(err)
			return &pb.Void{}, err
		}

		log.Infof("user physical home created at %s", pp)

		in := &proppb.GetReq{}
		in.Path = home
		in.AccessToken = req.AccessToken
		in.ForceCreation = true

		_, err = client.Get(ctx, in)
		if err != nil {
			return &pb.Void{}, nil
		}

		log.Info("home saved to %s", s.p.prop)

		return &pb.Void{}, nil
	}

	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("user physical home at %s already created")

	in := &proppb.GetReq{}
	in.Path = home
	in.AccessToken = req.AccessToken
	in.ForceCreation = true

	_, err = client.Get(ctx, in)
	if err != nil {
		return &pb.Void{}, nil
	}

	return &pb.Void{}, nil
}

func (s *server) Mkdir(ctx context.Context, req *pb.MkdirReq) (*pb.Void, error) {

	traceID, err := getTraceID(ctx)
	if err != nil {
		rus.Error(err)
		return &pb.Void{}, err
	}
	log := rus.WithField("trace", traceID).WithField("svc", serviceID)
	ctx = newTraceContext(ctx, traceID)

	log.Info("request started")

	// Time request
	reqStart := time.Now()

	defer func() {
		// Compute request duration
		reqDur := time.Since(reqStart)

		// Log access info
		log.WithFields(rus.Fields{
			"method":   "mkdir",
			"type":     "grpcaccess",
			"duration": reqDur.Seconds(),
		}).Info("request finished")

	}()

	idt, err := authlib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	log.Infof("%s", idt)

	p := path.Clean(req.Path)

	log.Infof("path is %s", p)

	if !isUnderHome(p, idt) {
		log.Error(permissionDenied)
		return &pb.Void{}, permissionDenied
	}

	if p == getHome(idt) {
		return &pb.Void{}, grpc.Errorf(codes.PermissionDenied, "cannot create directory")
	}

	pp := s.getPhysicalPath(p)

	log.Infof("physical path is %s", pp)

	err = os.Mkdir(pp, dirPerm)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("created dir %s", pp)

	resource, err := s.grpcPool.Get("")
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	defer resource.Release()

	handle, err := resource.Handle()
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}
	con := handle.(*grpc.ClientConn)
	log.Infof("created connection to %s", s.p.prop)

	client := proppb.NewPropClient(con)

	in := &proppb.PutReq{}
	in.Path = p
	in.AccessToken = req.AccessToken

	_, err = client.Put(ctx, in)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("dir %s added to prop", p)

	return &pb.Void{}, nil
}

func (s *server) Stat(ctx context.Context, req *pb.StatReq) (*pb.Metadata, error) {

	traceID, err := getTraceID(ctx)
	if err != nil {
		rus.Error(err)
		return &pb.Metadata{}, err
	}
	log := rus.WithField("trace", traceID).WithField("svc", serviceID)
	ctx = newTraceContext(ctx, traceID)

	log.Info("request started")

	// Time request
	reqStart := time.Now()

	defer func() {
		// Compute request duration
		reqDur := time.Since(reqStart)

		// Log access info
		log.WithFields(rus.Fields{
			"method":   "stat",
			"type":     "grpcaccess",
			"duration": reqDur.Seconds(),
		}).Info("request finished")

	}()

	idt, err := authlib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, unauthenticatedError
	}

	log.Infof("%s", idt)

	p := path.Clean(req.Path)

	log.Infof("path is %s", p)

	// The hierarchy is /local/users/d/demo
	// All paths in the hierarchy above the user home directory must be
	// accessible for all logged in users
	if isCommonDomain(p) {

	} else { // it must be under /local/users/{letter}
		if isUnderOtherHome(p, idt) {
			if req.Children == true {
				// TODO(labkode) Sharing
				return &pb.Metadata{}, permissionDenied
			}
		} else {
			// asset is under logged in user home directory
			if !isUnderHome(p, idt) {
				log.WithField("criticial", "").Errorf("path %s has not been handled correclty or fake path", p)
				return &pb.Metadata{}, permissionDenied
			}
		}
	}

	/*
		if !isUnderHome(p, idt) { // constrained to /local/users/d/demo/...
			log.Error(permissionDenied)
			return &pb.Metadata{}, permissionDenied
		}
	*/

	pp := s.getPhysicalPath(p)

	log.Infof("physical path is %s", pp)

	parentMeta, err := s.getMeta(pp)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	log.Infof("stated parent %s", pp)

	resource, err := s.grpcPool.Get("")
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	defer resource.Release()

	handle, err := resource.Handle()
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}
	con := handle.(*grpc.ClientConn)
	log.Infof("created connection to %s", s.p.prop)

	client := proppb.NewPropClient(con)

	in := &proppb.GetReq{}
	in.Path = p
	in.AccessToken = req.AccessToken
	in.ForceCreation = true

	rec, err := client.Get(ctx, in)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	parentMeta.Id = rec.Id
	parentMeta.Etag = rec.Etag
	parentMeta.Modified = rec.Modified
	parentMeta.Checksum = rec.Checksum

	if !parentMeta.IsContainer || req.Children == false {
		return parentMeta, nil
	}

	dir, err := os.Open(pp)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	log.Infof("opened dir %s", pp)

	defer dir.Close()

	names, err := dir.Readdirnames(0)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	log.Infof("dir %s has %d entries", pp, len(names))

	for _, n := range names {
		cp := path.Join(parentMeta.Path, path.Clean(n))
		cpp := s.getPhysicalPath(cp)
		m, err := s.getMeta(cpp)
		if err != nil {
			log.Error(err)
		} else {
			in := &proppb.GetReq{}
			in.Path = cp
			in.AccessToken = req.AccessToken
			in.ForceCreation = true

			rec, err := client.Get(ctx, in)
			if err != nil {
				log.Errorf("path %s has not been added because %s", p, err.Error())
			} else {
				m.Id = rec.Id
				m.Etag = rec.Etag
				m.Modified = rec.Modified
				m.Checksum = rec.Checksum
				parentMeta.Children = append(parentMeta.Children, m)

				log.Infof("added %s to parent", m.Path)
			}
		}
	}

	log.Infof("added %d entries to parent", len(parentMeta.Children))

	return parentMeta, nil
}

func (s *server) Cp(ctx context.Context, req *pb.CpReq) (*pb.Void, error) {

	traceID, err := getTraceID(ctx)
	if err != nil {
		rus.Error(err)
		return &pb.Void{}, err
	}
	log := rus.WithField("trace", traceID).WithField("svc", serviceID)
	ctx = newTraceContext(ctx, traceID)

	log.Info("request started")

	// Time request
	reqStart := time.Now()

	defer func() {
		// Compute request duration
		reqDur := time.Since(reqStart)

		// Log access info
		log.WithFields(rus.Fields{
			"method":   "cp",
			"type":     "grpcaccess",
			"duration": reqDur.Seconds(),
		}).Info("request finished")

	}()

	idt, err := authlib.ParseToken(req.AccessToken, s.p.sharedSecret)

	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	log.Infof("%s", idt)

	src := path.Clean(req.Src)
	dst := path.Clean(req.Dst)

	log.Infof("src is %s", src)
	log.Info("dst is %s", dst)

	if !isUnderHome(src, idt) {
		log.Error(permissionDenied)
		return &pb.Void{}, permissionDenied
	}

	if !isUnderHome(dst, idt) {
		log.Error(permissionDenied)
		return &pb.Void{}, permissionDenied
	}

	if src == getHome(idt) || dst == getHome(idt) {
		return &pb.Void{}, grpc.Errorf(codes.PermissionDenied, "cannot copy from/to home directory")
	}

	psrc := s.getPhysicalPath(src)
	pdst := s.getPhysicalPath(dst)

	log.Infof("physical src is %s", psrc)
	log.Infof("physical dst is %s", pdst)

	statReq := &pb.StatReq{}
	statReq.AccessToken = req.AccessToken
	statReq.Path = req.Src

	meta, err := s.Stat(ctx, statReq)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("stated %s", src)

	if meta.IsContainer {
		err = copyDir(psrc, pdst)
		if err != nil {
			log.Error(err)
			return &pb.Void{}, err
		}

		log.Infof("copied from dir %s to dir %s", psrc, pdst)
	} else {
		// It is a file
		err = copyFile(psrc, pdst, int64(meta.Size))
		if err != nil {
			log.Error(err)
			return &pb.Void{}, err
		}

		log.Infof("copied from file %s to file %s", psrc, pdst)
	}

	resource, err := s.grpcPool.Get("")
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	defer resource.Release()

	handle, err := resource.Handle()
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}
	con := handle.(*grpc.ClientConn)
	log.Infof("created connection to %s", s.p.prop)

	client := proppb.NewPropClient(con)

	in := &proppb.PutReq{}
	in.Path = dst
	in.AccessToken = req.AccessToken

	_, err = client.Put(ctx, in)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("copied resource %s saved in prop")

	return &pb.Void{}, nil
}

func (s *server) Mv(ctx context.Context, req *pb.MvReq) (*pb.Void, error) {

	traceID, err := getTraceID(ctx)
	if err != nil {
		rus.Error(err)
		return &pb.Void{}, err
	}
	log := rus.WithField("trace", traceID).WithField("svc", serviceID)
	ctx = newTraceContext(ctx, traceID)

	log.Info("request started")

	// Time request
	reqStart := time.Now()

	defer func() {
		// Compute request duration
		reqDur := time.Since(reqStart)

		// Log access info
		log.WithFields(rus.Fields{
			"method":   "mv",
			"type":     "grpcaccess",
			"duration": reqDur.Seconds(),
		}).Info("request finished")

	}()

	idt, err := authlib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	log.Infof("%s", idt)

	src := path.Clean(req.Src)
	dst := path.Clean(req.Dst)

	log.Infof("src is %s", src)
	log.Info("dst is %s", dst)

	if !isUnderHome(src, idt) {
		log.Error(permissionDenied)
		return &pb.Void{}, permissionDenied
	}

	if !isUnderHome(dst, idt) {
		log.Error(permissionDenied)
		return &pb.Void{}, permissionDenied
	}

	if src == getHome(idt) || dst == getHome(idt) {
		return &pb.Void{}, grpc.Errorf(codes.PermissionDenied, "cannot rename from/to home directory")
	}

	psrc := s.getPhysicalPath(src)
	pdst := s.getPhysicalPath(dst)

	log.Infof("physical src is %s", psrc)
	log.Infof("physical dst is %s", pdst)

	err = os.Rename(psrc, pdst)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("renamed from %s to %s", psrc, pdst)

	resource, err := s.grpcPool.Get("")
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	defer resource.Release()

	handle, err := resource.Handle()
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}
	con := handle.(*grpc.ClientConn)
	log.Infof("created connection to %s", s.p.prop)

	client := proppb.NewPropClient(con)

	in := &proppb.MvReq{}
	in.Src = src
	in.Dst = dst
	in.AccessToken = req.AccessToken

	_, err = client.Mv(ctx, in)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("renamed %s to %s in prop", src, dst)

	return &pb.Void{}, nil
}

func (s *server) Rm(ctx context.Context, req *pb.RmReq) (*pb.Void, error) {

	traceID, err := getTraceID(ctx)
	if err != nil {
		rus.Error(err)
		return &pb.Void{}, err
	}
	log := rus.WithField("trace", traceID).WithField("svc", serviceID)
	ctx = newTraceContext(ctx, traceID)

	log.Info("request started")

	// Time request
	reqStart := time.Now()

	defer func() {
		// Compute request duration
		reqDur := time.Since(reqStart)

		// Log access info
		log.WithFields(rus.Fields{
			"method":   "rm",
			"type":     "grpcaccess",
			"duration": reqDur.Seconds(),
		}).Info("request finished")

	}()

	idt, err := authlib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	log.Infof("%s", idt)

	p := path.Clean(req.Path)

	log.Infof("path is %s", p)

	if !isUnderHome(p, idt) {
		log.Error(permissionDenied)
		return &pb.Void{}, permissionDenied
	}

	if p == getHome(idt) {
		return &pb.Void{}, grpc.Errorf(codes.PermissionDenied, "cannot remove home directory")
	}

	pp := s.getPhysicalPath(p)

	log.Infof("physical path is %s", pp)

	err = os.RemoveAll(pp)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("removed %s", pp)

	resource, err := s.grpcPool.Get("")
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	defer resource.Release()

	handle, err := resource.Handle()
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}
	con := handle.(*grpc.ClientConn)
	log.Infof("created connection to %s", s.p.prop)
	client := proppb.NewPropClient(con)

	in := &proppb.RmReq{}
	in.Path = p
	in.AccessToken = req.AccessToken

	_, err = client.Rm(ctx, in)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	log.Infof("paths with prefix %s removed from prop", p)

	return &pb.Void{}, nil
}

// getMeta return the metadata of path pp.
func (s *server) getMeta(pp string) (*pb.Metadata, error) {

	finfo, err := os.Stat(pp)
	if err != nil {
		return &pb.Metadata{}, err
	}

	logicalPath := s.getLogicalPath(pp)

	m := &pb.Metadata{}
	m.Path = logicalPath
	m.Size = uint32(finfo.Size())
	m.IsContainer = finfo.IsDir()
	m.Permissions = 0
	m.MimeType = mime.TypeByExtension(path.Ext(m.Path))

	if m.MimeType == "" {
		m.MimeType = "application/octet-stream"
	}

	if m.IsContainer {
		m.MimeType = "inode/container"
	}

	return m, nil
}

func (s *server) getPhysicalPath(p string) string {
	return path.Join(s.p.dataDir, path.Clean(p))
}

func (s *server) getLogicalPath(pp string) string {
	return path.Clean(strings.TrimPrefix(pp, s.p.dataDir))
}
