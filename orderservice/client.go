package orderservice

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	orderServiceClientAddress string
)

func init() {
	flag.StringVar(&orderServiceClientAddress, "orderservice-address", "orderservice.orders", "")
}

func NewClient(ctx context.Context) (protos.OrderServiceClient, error) {
	logrus.WithField("address", orderServiceClientAddress).Debug("connecting to order service")
	conn, err := grpc.DialContext(ctx, orderServiceClientAddress, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	logrus.WithField("address", orderServiceClientAddress).Debug("connected to order service")
	return protos.NewOrderServiceClient(conn), nil
}
