package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

var (
	level           string
	serverAddress   string
	databaseAddress string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&serverAddress, "server-address", ":8080", "")
	flag.StringVar(&databaseAddress, "database-address", "", "")
}

func main() {
	flag.Parse()
	if ll, err := logrus.ParseLevel(level); err != nil {
		logrus.WithError(err).Fatal("parsing log level")
	} else {
		logrus.SetLevel(ll)
	}

	ctx := context.Background()

	server, err := dataservice.NewServer(ctx, serverAddress)
	if err != nil {
		logrus.WithError(err).Fatal("creating server")
	}

	grpcServer := grpc.NewServer()
	protos.RegisterDataServiceServer(grpcServer, server)

	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		logrus.WithError(err).Fatal("listening tcp")
	}

	if err := grpcServer.Serve(listener); err != nil {
		logrus.WithError(err).Fatal("serving grpc")
	}
}
