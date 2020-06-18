package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/pubsub"
	"github.com/ducc/kwɒnt/tick_writer"
	"github.com/sirupsen/logrus"
)

var (
	level string
	topic string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
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

	tick_writer.Run(ctx, ds, msgs)
}
