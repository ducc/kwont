package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/tick_writer"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

var (
	level        string
	natsAddress  string
	natsUsername string
	natsPassword string
	topic        string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&natsAddress, "nats-address", "127.0.0.1:4150", "nats server address")
	flag.StringVar(&natsUsername, "nats-username", "kwont", "nats username")
	flag.StringVar(&natsPassword, "nats-password", "password", "nats password")
	flag.StringVar(&topic, "topic", "", "")
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

	natsConn, err := nats.Connect(natsAddress, nats.UserInfo(natsUsername, natsPassword))
	if err != nil {
		logrus.WithError(err).Fatal("connecting to nats")
	}

	subscription, err := natsConn.SubscribeSync(topic)
	if err != nil {
		logrus.WithError(err).Fatal("subscribing to topic")
	}

	tick_writer.Run(ctx, ds, subscription)
}
