package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/brokers"
	"github.com/ducc/kwɒnt/brokers/xtb"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	level         string
	amqpAddress   string
	topic         string
	serverAddress string
	routerAddress string
)

func init() {
	flag.StringVar(&level, "level", "debug", "logrus logging level")
	flag.StringVar(&amqpAddress, "amqp-address", "", "amqp server connection address")
	flag.StringVar(&topic, "topic", "ticks", "nats topic")
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

	amqpConn, err := amqp.Dial(amqpAddress)
	if err != nil {
		logrus.WithError(err).Fatal("connecting to amqp server")
	}
	defer func() {
		if err := amqpConn.Close(); err != nil {
			logrus.WithError(err).Error("closing amqp conn")
		}
	}()

	amqpChan, err := amqpConn.Channel()
	if err != nil {
		logrus.WithError(err).Fatal("creating amqp channel")
	}
	defer func() {
		if err := amqpChan.Close(); err != nil {
			logrus.WithError(err).Error("closing amqp chan")
		}
	}()

	amqpQueue, err := amqpChan.QueueDeclare(
		topic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.WithError(err).Fatal("declaring amqp queue")
	}

	ctx := context.Background()

	routerConn, err := brokers.NewClient(ctx, routerAddress)
	if err != nil {
		logrus.WithError(err).Fatal("connecting to router")
	}

	server := xtb.New(amqpChan, amqpQueue, topic, routerConn)
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
