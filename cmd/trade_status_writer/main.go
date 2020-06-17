package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/ducc/kwɒnt/trade_status_writer"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	level       string
	amqpAddress string
	topic       string
	broker      string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&amqpAddress, "amqp-address", "", "amqp server connection address")
	flag.StringVar(&topic, "topic", "", "")
	flag.StringVar(&broker, "broker", "", "")
}

func main() {
	flag.Parse()
	if ll, err := logrus.ParseLevel(level); err != nil {
		logrus.WithError(err).Fatal("parsing log level")
	} else {
		logrus.SetLevel(ll)
	}

	brokerName := protos.Broker_Name(protos.Broker_Name_value[broker])
	if brokerName == protos.Broker_UNKNOWN {
		logrus.Fatal("unknown broker")
	}

	ctx := context.Background()

	ds, err := dataservice.NewClient(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("creating dataservice client")
	}

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

	q, err := amqpChan.QueueDeclare(
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

	msgs, err := amqpChan.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.WithError(err).Fatal("declaring amqp consumer")
	}

	trade_status_writer.Run(ctx, ds, brokerName, msgs)
}
