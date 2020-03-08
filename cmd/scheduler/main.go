package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/scheduler"
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
	flag.StringVar(&topic, "topic", "", "")
	flag.StringVar(&natsAddress, "nats-address", "127.0.0.1:4150", "nats server address")
	flag.StringVar(&natsUsername, "nats-username", "kwont", "nats username")
	flag.StringVar(&natsPassword, "nats-password", "password", "nats password")
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

	scheduler.Run(ctx, ds, natsConn, topic)
}
