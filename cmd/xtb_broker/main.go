package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/brokers"
	"github.com/ducc/kwɒnt/brokers/xtb"
	"github.com/ducc/kwɒnt/protos"
	"github.com/ducc/kwɒnt/pubsub"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	level            string
	tickTopic        string
	tradeTopic       string
	tradeStatusTopic string
	serverAddress    string
	routerAddress    string
)

func init() {
	flag.StringVar(&level, "level", "debug", "logrus logging level")
	flag.StringVar(&tickTopic, "tick-topic", "ticks", "nats topic")
	flag.StringVar(&tradeTopic, "trade-topic", "xtb-trades", "nats topic")
	flag.StringVar(&tradeStatusTopic, "trade-status-topic", "xtb-trade-status", "nats topic")
	flag.StringVar(&serverAddress, "server-address", ":8080", "grpc server address")
	flag.StringVar(&routerAddress, "router-address", "", "router service address")
}

func main() {
	flag.Parse()
	if ll, err := logrus.ParseLevel(level); err != nil {
		logrus.WithError(err).Fatal("parsing log level")
	} else {
		logrus.SetLevel(ll)
	}

	logrus.WithField("POD_IP", os.Getenv("POD_IP")).Debug("starting xtb broker")

	psClient, err := pubsub.New()
	if err != nil {
		logrus.WithError(err).Fatal("creating pubsub client")
	}
	defer func() {
		if err := psClient.Close(); err != nil {
			logrus.WithError(err).Error("closing pubsub client")
		}
	}()

	tickQueue, err := psClient.Queue(tickTopic)
	if err != nil {
		logrus.WithError(err).Fatal("creating pubsub queue")
	}

	tradeQueue, err := psClient.Queue(tradeTopic)
	if err != nil {
		logrus.WithError(err).Fatal("creating pubsub queue")
	}

	tradeStatusQueue, err := psClient.Queue(tradeStatusTopic)
	if err != nil {
		logrus.WithError(err).Fatal("creating pubsub queue")
	}

	ctx := context.Background()

	routerConn, err := brokers.NewClient(ctx, routerAddress)
	if err != nil {
		logrus.WithError(err).Fatal("connecting to router")
	}

	server := xtb.New(tickQueue, tradeQueue, tradeStatusQueue, routerConn)
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
