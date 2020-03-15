package brokers

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

func NewClient(ctx context.Context, address string) (protos.BrokerServiceClient, error) {
	logrus.WithField("address", address).Debug("connecting to broker service")

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return protos.NewBrokerServiceClient(conn), nil
}
