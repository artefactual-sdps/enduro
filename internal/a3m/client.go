package a3m

import (
	context "context"

	"buf.build/gen/go/artefactual/a3m/grpc/go/a3m/api/transferservice/v1beta1/transferservicev1beta1grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a client of a3m that remembers and reuses the underlying gPRC client.
type Client struct {
	TransferClient transferservicev1beta1grpc.TransferServiceClient
}

var currClient *Client

func NewClient(ctx context.Context, addr string) (*Client, error) {
	if currClient != nil {
		// Do we need to call conn.Connect()?
		return currClient, nil
	}

	c := &Client{}

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	currClient = c
	c.TransferClient = transferservicev1beta1grpc.NewTransferServiceClient(conn)

	return c, nil
}
