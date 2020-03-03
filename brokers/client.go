package brokers

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"google.golang.org/grpc"
)

func NewClient(ctx context.Context, address string) (protos.BrokerServiceClient, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return protos.NewBrokerServiceClient(conn), nil
}
