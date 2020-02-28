package dataservice

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/protos"
	"google.golang.org/grpc"
)

var (
	dataServiceClientAddress string
)

func init() {
	flag.StringVar(&dataServiceClientAddress, "data-service-client-address", "", "") // todo default
}

func NewClient(ctx context.Context) (protos.DataServiceClient, error) {
	conn, err := grpc.DialContext(ctx, dataServiceClientAddress, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return protos.NewDataServiceClient(conn), nil
}
