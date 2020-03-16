package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/brokers/router"
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
	flag.StringVar(&level, "level", "debug", "logrus logging level")
	flag.StringVar(&serverAddress, "server-address", ":8080", "grpc server address")
}

func main() {
	flag.Parse()
	if ll, err := logrus.ParseLevel(level); err != nil {
		logrus.WithError(err).Fatal("parsing log level")
	} else {
		logrus.SetLevel(ll)
	}

	logrus.Debug("starting router")

	ctx := context.Background()

	server := router.NewRouter()
	grpcServer := grpc.NewServer()

	protos.RegisterBrokerServiceServer(grpcServer, server)

	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		logrus.WithError(err).Fatal("tcp listen")
	}

	if err := grpcServer.Serve(listener); err != nil {
		logrus.WithError(err).Fatal("serving grpc")
	}
}
