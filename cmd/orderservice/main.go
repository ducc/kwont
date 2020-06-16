package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/brokers"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/orderservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

var (
	level         string
	serverAddress string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&serverAddress, "server-address", ":8080", "")
}

func main() {
	flag.Parse()
	if ll, err := logrus.ParseLevel(level); err != nil {
		logrus.WithError(err).Fatal("parsing log level")
	} else {
		logrus.SetLevel(ll)
	}

	ctx := context.Background()

	ds, err := dataservice.NewClient(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("creating ds client")
	}

	broker, err := brokers.NewClient(ctx, "") // todo remove address param as there is a flag already
	if err != nil {
		logrus.WithError(err).Fatal("creating broker client")
	}

	server, err := orderservice.NewServer(ctx, ds, broker)
	if err != nil {
		logrus.WithError(err).Fatal("creating server")
	}

	grpcServer := grpc.NewServer()
	protos.RegisterOrderServiceServer(grpcServer, server)

	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		logrus.WithError(err).Fatal("listening tcp")
	}

	if err := grpcServer.Serve(listener); err != nil {
		logrus.WithError(err).Fatal("serving grpc")
	}
}
