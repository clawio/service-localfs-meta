package main

import (
	pb "github.com/clawio/service.localstore.meta/proto"
	"golang.org/x/net/context"
	"os"
	"path"
)

const (
	dirPerm = 0755
)

type newServerParams struct {
	dataDir string
	tmpDir  string
}

func newServer(p *newServerParams) *server {
	return &server{p}
}

type server struct {
	p *newServerParams
}

func (s *server) getHome(idt *pb.Identity) string {
	return path.Join(s.p.dataDir, path.Join(idt.Pid))
}

func (s *server) Home(ctx context.Context, req *pb.HomeReq) (*pb.Void, error) {
	home := s.getHome(req.Idt)
	_, err := os.Stat(home)

	// Create home dir if not exists
	if os.IsNotExist(err) {
		err = os.Mkdir(home, dirPerm)
		if err != nil {
			return nil, err
		}
		return &pb.Void{}, nil
	}

	if err != nil {
		return nil, err
	}

	return &pb.Void{}, nil
}

func (s *server) Mkdir(ctx context.Context, req *pb.MkdirReq) (*pb.Void, error) {
	p := path.Join(s.getHome(req.Idt), path.Clean(req.Path))
	return &pb.Void{}, os.Mkdir(p, dirPerm)
}

func (s *server) Stat(ctx context.Context, req *pb.StatReq) (*pb.Metadata, error) {
	p := path.Join(s.getHome(req.Idt), path.Clean(req.Path))

	finfo, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	m := &pb.Metadata{}
	m.Id = "TODO"
	m.Path = path.Clean(p)
	m.Size = uint32(finfo.Size())
	m.IsContainer = finfo.IsDir()
	m.Modified = uint32(finfo.ModTime().Unix())
	m.Etag = "TODO"
	m.Permissions = 0

	return m, nil
}

func (s *server) Cp(ctx context.Context, req *pb.CpReq) (*pb.Void, error) {
	src := path.Join(s.getHome(req.Idt), path.Clean(req.Src))
	dst := path.Join(s.getHome(req.Idt), path.Clean(req.Dst))

	statReq := &pb.StatReq{}
	statReq.Idt = req.Idt
	statReq.Path = req.Src

	meta, err := s.Stat(ctx, statReq)
	if err != nil {
		return &pb.Void{}, err
	}

	if meta.IsContainer {
		return &pb.Void{}, copyDir(src, dst)
	}

	return &pb.Void{}, copyFile(src, dst, int64(meta.Size))
}

func (s *server) Mv(ctx context.Context, req *pb.MvReq) (*pb.Void, error) {
	src := path.Join(s.getHome(req.Idt), path.Clean(req.Src))
	dst := path.Join(s.getHome(req.Idt), path.Clean(req.Dst))

	return &pb.Void{}, os.Rename(src, dst)
}

func (s *server) Rm(ctx context.Context, req *pb.RmReq) (*pb.Void, error) {
	p := path.Join(s.getHome(req.Idt), path.Clean(req.Path))

	return &pb.Void{}, os.Remove(p)
}
