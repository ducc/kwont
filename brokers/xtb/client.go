package xtb

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/protos"
	"google.golang.org/grpc"
)

var (
	brokerClientAddress string
)

func init() {
	flag.StringVar(&brokerClientAddress, "broker-client-address", "", "") // todo default
}

func NewClient(ctx context.Context) (protos.BrokerServiceClient, error) {
	conn, err := grpc.DialContext(ctx, brokerClientAddress, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return protos.NewBrokerServiceClient(conn), nil
}
