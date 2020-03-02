package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/price_writer"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

var (
	level   string
	address string
	channel string
	topic   string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&address, "address", "", "")
	flag.StringVar(&channel, "channel", "", "")
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

	config := nsq.NewConfig()
	// todo configure

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		logrus.WithError(err).Fatal("creating consumer")
	}

	if err := consumer.ConnectToNSQLookupd(address); err != nil {
		logrus.WithError(err).Fatal("connecting to nsq")
	}

	price_writer.Run(ctx, ds, consumer, topic)
}
