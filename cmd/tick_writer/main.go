package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/tick_writer"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	level       string
	amqpAddress string
	topic       string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&amqpAddress, "amqp-address", "", "amqp server connection address")
	flag.StringVar(&topic, "topic", "ticks", "")
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

	tick_writer.Run(ctx, ds, msgs)
}
