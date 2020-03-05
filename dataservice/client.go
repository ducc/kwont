package dataservice

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	dataServiceClientAddress string
)

func init() {
	flag.StringVar(&dataServiceClientAddress, "dataservice-address", "dataservice.data", "")
}

func NewClient(ctx context.Context) (protos.DataServiceClient, error) {
	logrus.WithField("address", dataServiceClientAddress).Debug("connecting to data service")
	conn, err := grpc.DialContext(ctx, dataServiceClientAddress, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	logrus.WithField("address", dataServiceClientAddress).Debug("connected to data service")
	return protos.NewDataServiceClient(conn), nil
}
