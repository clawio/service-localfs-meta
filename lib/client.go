package lib

import (
	pb "github.com/clawio/service.localstore.meta/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewClientParams are the params used to create a client.
type NewClientParams struct {
	Addr string
	Opts []grpc.DialOption
}

// NewClient create a client or returns an error.
func NewClient(p *NewClientParams) (*Client, error) {
	conn, err := grpc.Dial(p.Addr, p.Opts...)
	if err != nil {
		return nil, err
	}

	client := pb.NewLocalClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil

}

// Client handles the operations agains the gRPC/HTTP server.
type Client struct {
	conn   *grpc.ClientConn
	client pb.LocalClient
}

// Mkdir creates a directory.
func (c *Client) Mkdir(ctx context.Context, r *pb.MkdirReq) (*pb.Void, error) {
	return c.client.Mkdir(ctx, r)
}

// Home creates the identity home directory.
func (c *Client) Home(ctx context.Context, r *pb.HomeReq) (*pb.Void, error) {
	return c.client.Home(ctx, r)
}

// Stat stats a resource or returns an error. If children is set to true,
// direct children will be included in the response.
func (c *Client) Stat(ctx context.Context, r *pb.StatReq) (*pb.Metadata, error) {
	return c.client.Stat(ctx, r)
}

// Cp copies a resource from src to dst or returns an error.
func (c *Client) Cp(ctx context.Context, r *pb.CpReq) (*pb.Void, error) {
	return c.client.Cp(ctx, r)
}

// Mv moves a resource from src to dst or returns an error.
func (c *Client) Mv(ctx context.Context, r *pb.MvReq) (*pb.Void, error) {
	return c.client.Mv(ctx, r)
}

// Rm removes a resource or returns an error.
func (c *Client) Rm(ctx context.Context, r *pb.RmReq) (*pb.Void, error) {
	return c.client.Rm(ctx, r)
}

// Close closes the client connection.
func (c Client) Close() error {
	return c.conn.Close()
}
