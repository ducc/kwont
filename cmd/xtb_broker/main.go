package main

import (
	"flag"
	"github.com/ducc/kwɒnt/brokers/xtb"
	"github.com/ducc/kwɒnt/protos"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

var (
	level         string
	username      string
	password      string
	natsAddress   string
	natsUsername  string
	natsPassword  string
	topic         string
	serverAddress string
)

func init() {
	flag.StringVar(&level, "level", "debug", "logrus logging level")
	flag.StringVar(&username, "username", "", "xopen hub username")
	flag.StringVar(&password, "password", "", "xopen hub password")
	flag.StringVar(&natsAddress, "nats-address", "127.0.0.1:4150", "nats server address")
	flag.StringVar(&natsUsername, "nats-username", "kwont", "nats username")
	flag.StringVar(&natsPassword, "nats-password", "password", "nats password")
	flag.StringVar(&topic, "topic", "candlesticks", "nats topic")
	flag.StringVar(&serverAddress, "server-address", ":8080", "grpc server address")
}

func main() {
	flag.Parse()
	if ll, err := logrus.ParseLevel(level); err != nil {
		logrus.WithError(err).Fatal("parsing log level")
	} else {
		logrus.SetLevel(ll)
	}

	natsConn, err := nats.Connect(natsAddress, nats.UserInfo(natsUsername, natsPassword))
	if err != nil {
		logrus.WithError(err).Fatal("connecting to nats")
	}

	server := xtb.New(natsConn, topic)
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
