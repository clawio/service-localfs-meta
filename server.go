package main

import (
	"github.com/clawio/service.auth/lib"
	pb "github.com/clawio/service.localstore.meta/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"os"
	"path"
)

const (
	dirPerm = 0755
)

var (
	unauthenticatedError = grpc.Errorf(codes.Unauthenticated, "identity not found")
)

type newServerParams struct {
	dataDir      string
	tmpDir       string
	sharedSecret string
}

func newServer(p *newServerParams) *server {
	return &server{p}
}

type server struct {
	p *newServerParams
}

func (s *server) getHome(idt *lib.Identity) string {
	return path.Join(s.p.dataDir, path.Join(idt.Pid))
}

func (s *server) Home(ctx context.Context, req *pb.HomeReq) (*pb.Void, error) {

	idt, err := lib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	home := s.getHome(idt)
	_, err = os.Stat(home)

	// Create home dir if not exists
	if os.IsNotExist(err) {
		err = os.Mkdir(home, dirPerm)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		return &pb.Void{}, nil
	}

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &pb.Void{}, nil
}

func (s *server) Mkdir(ctx context.Context, req *pb.MkdirReq) (*pb.Void, error) {

	idt, err := lib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	p := path.Join(s.getHome(idt), path.Clean(req.Path))
	return &pb.Void{}, os.Mkdir(p, dirPerm)
}

func (s *server) Stat(ctx context.Context, req *pb.StatReq) (*pb.Metadata, error) {

	idt, err := lib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, unauthenticatedError
	}

	p := path.Join(s.getHome(idt), path.Clean(req.Path))

	parentMeta, err := s.getMeta(p)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	if !parentMeta.IsContainer || req.Children == false {
		return parentMeta, nil
	}

	dir, err := os.Open(p)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	defer dir.Close()

	names, err := dir.Readdirnames(0)
	if err != nil {
		log.Error(err)
		return &pb.Metadata{}, err
	}

	for _, n := range names {

		m, err := s.getMeta(n)
		if err != nil {
			log.Error(err)
		}

		parentMeta.Children = append(parentMeta.Children, m)
	}

	return parentMeta, nil
}

func (s *server) Cp(ctx context.Context, req *pb.CpReq) (*pb.Void, error) {

	idt, err := lib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	src := path.Join(s.getHome(idt), path.Clean(req.Src))
	dst := path.Join(s.getHome(idt), path.Clean(req.Dst))

	statReq := &pb.StatReq{}
	statReq.AccessToken = req.AccessToken
	statReq.Path = req.Src

	meta, err := s.Stat(ctx, statReq)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, err
	}

	if meta.IsContainer {
		return &pb.Void{}, copyDir(src, dst)
	}

	return &pb.Void{}, copyFile(src, dst, int64(meta.Size))
}

func (s *server) Mv(ctx context.Context, req *pb.MvReq) (*pb.Void, error) {

	idt, err := lib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	src := path.Join(s.getHome(idt), path.Clean(req.Src))
	dst := path.Join(s.getHome(idt), path.Clean(req.Dst))

	return &pb.Void{}, os.Rename(src, dst)
}

func (s *server) Rm(ctx context.Context, req *pb.RmReq) (*pb.Void, error) {

	idt, err := lib.ParseToken(req.AccessToken, s.p.sharedSecret)
	if err != nil {
		log.Error(err)
		return &pb.Void{}, unauthenticatedError
	}

	p := path.Join(s.getHome(idt), path.Clean(req.Path))

	return &pb.Void{}, os.Remove(p)
}
