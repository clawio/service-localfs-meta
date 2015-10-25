package lib

import (
	pb "github.com/clawio/service.localstore.meta/proto"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	. "gopkg.in/check.v1"
	"path"
	"testing"
)

const Addr = "localhost:57001"

func Test(t *testing.T) { TestingT(t) }

type S struct {
	c   *Client
	idt *pb.Identity
}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
	p := &NewClientParams{}
	p.Addr = Addr
	p.Opts = []grpc.DialOption{grpc.WithInsecure()}

	client, err := NewClient(p)
	if err != nil {
		c.Fatal(err)
	}

	s.c = client

	idt := &pb.Identity{}
	idt.Pid = "hugo"
	idt.Idp = "localhost"
	idt.DisplayName = "Hugo González Labrador"

	s.idt = idt
}

func (s *S) TestHome(c *C) {
	ctx := context.Background()

	r := &pb.HomeReq{}
	r.Idt = s.idt

	_, err := s.c.Home(ctx, r)
	if err != nil {
		c.Fatal(err)
	}
}

func (s *S) TestMkdir(c *C) {
	ctx := context.Background()

	ran, err := uuid.NewV4()
	if err != nil {
		c.Fatal(err)
	}

	r := &pb.MkdirReq{}
	r.Idt = s.idt
	r.Path = ran.String()

	_, err = s.c.Mkdir(ctx, r)
	if err != nil {
		c.Fatal(err)
	}
}

func (s *S) TestMkdirAlreadyExists(c *C) {
	ctx := context.Background()

	ran, err := uuid.NewV4()
	if err != nil {
		c.Fatal(err)
	}

	r := &pb.MkdirReq{}
	r.Idt = s.idt
	r.Path = ran.String()

	_, err = s.c.Mkdir(ctx, r)
	if err != nil {
		c.Fatal(err)
	}
	_, err = s.c.Mkdir(ctx, r)
	if grpc.Code(err) != codes.AlreadyExists {
		c.Fatal(err)
	}
}

func (s *S) TestMkdirParentNotFound(c *C) {
	ctx := context.Background()

	ran, err := uuid.NewV4()
	if err != nil {
		c.Fatal(err)
	}

	r := &pb.MkdirReq{}
	r.Idt = s.idt
	r.Path = path.Join("notexists", ran.String())

	_, err = s.c.Mkdir(ctx, r)
	if grpc.Code(err) != codes.NotFound {
		c.Fatal(err)
	}
}

func (s *S) TestCpDir(c *C) {
	ctx := context.Background()

	ran, err := uuid.NewV4()
	if err != nil {
		c.Fatal(err)
	}

	mkdirReq := &pb.MkdirReq{}
	mkdirReq.Idt = s.idt
	mkdirReq.Path = ran.String()

	_, err = s.c.Mkdir(ctx, mkdirReq)
	if err != nil {
		c.Error(err)
	}

	r := &pb.CpReq{}
	r.Idt = s.idt
	r.Src = ran.String()
	r.Dst = ran.String() + "-dst"

	_, err = s.c.Cp(ctx, r)
	if err != nil {
		c.Fatal(err)
	}
}

// TODO(labkode) Use s.Upload when implemented to create tmp file
func (s *S) TestCpFile(c *C) {
	ctx := context.Background()

	ran, err := uuid.NewV4()
	if err != nil {
		c.Fatal(err)
	}

	mkdirReq := &pb.MkdirReq{}
	mkdirReq.Idt = s.idt
	mkdirReq.Path = ran.String()

	_, err = s.c.Mkdir(ctx, mkdirReq)
	if err != nil {
		c.Error(err)
	}

	r := &pb.CpReq{}
	r.Idt = s.idt
	r.Src = ran.String()
	r.Dst = ran.String() + "-dst"

	_, err = s.c.Cp(ctx, r)
	if err != nil {
		c.Fatal(err)
	}
}

func (s *S) TestMv(c *C) {
	ctx := context.Background()

	ran, err := uuid.NewV4()
	if err != nil {
		c.Fatal(err)
	}

	mkdirReq := &pb.MkdirReq{}
	mkdirReq.Idt = s.idt
	mkdirReq.Path = ran.String()

	_, err = s.c.Mkdir(ctx, mkdirReq)
	if err != nil {
		c.Error(err)
	}

	r := &pb.MvReq{}
	r.Idt = s.idt
	r.Src = ran.String()
	r.Dst = ran.String() + "-dst"

	_, err = s.c.Mv(ctx, r)
	if err != nil {
		c.Fatal(err)
	}

}

func (s *S) TestRm(c *C) {
	ctx := context.Background()

	ran, err := uuid.NewV4()
	if err != nil {
		c.Fatal(err)
	}

	mkdirReq := &pb.MkdirReq{}
	mkdirReq.Idt = s.idt
	mkdirReq.Path = ran.String()

	_, err = s.c.Mkdir(ctx, mkdirReq)
	if err != nil {
		c.Error(err)
	}

	r := &pb.RmReq{}
	r.Idt = s.idt
	r.Path = ran.String()

	_, err = s.c.Rm(ctx, r)
	if err != nil {
		c.Fatal(err)
	}

	statReq := &pb.StatReq{}
	statReq.Idt = s.idt
	statReq.Path = ran.String()

	_, err = s.c.Stat(ctx, statReq)
	if grpc.Code(err) != codes.NotFound {
		c.Fatal(err)
	}
}
