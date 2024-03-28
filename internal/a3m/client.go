package a3m

import (
	context "context"

	"buf.build/gen/go/artefactual/a3m/grpc/go/a3m/api/transferservice/v1beta1/transferservicev1beta1grpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client of a3m using the gRPC API.
type Client struct {
	TransferClient transferservicev1beta1grpc.TransferServiceClient
}

func NewClient(ctx context.Context, tp trace.TracerProvider, addr string) (*Client, error) {
	c := &Client{}

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(
			otelgrpc.NewClientHandler(
				otelgrpc.WithTracerProvider(tp),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	c.TransferClient = transferservicev1beta1grpc.NewTransferServiceClient(conn)

	return c, nil
}
