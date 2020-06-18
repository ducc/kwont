package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/ducc/kwɒnt/pubsub"
	"github.com/ducc/kwɒnt/trade_writer"
	"github.com/sirupsen/logrus"
)

var (
	level  string
	topic  string
	broker string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
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

	psClient, err := pubsub.New()
	if err != nil {
		logrus.WithError(err).Fatal("creating pubsub client")
	}
	defer func() {
		if err := psClient.Close(); err != nil {
			logrus.WithError(err).Error("closing pubsub client")
		}
	}()

	psQueue, err := psClient.Queue(topic)
	if err != nil {
		logrus.WithError(err).Fatal("creating pubsub queue")
	}

	msgs, err := psQueue.Subscribe()
	if err != nil {
		logrus.WithError(err).Fatal("subscribing to pubsub queue")
	}

	trade_writer.Run(ctx, ds, brokerName, msgs)
}
