package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/scheduler"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

var (
	level           string
	producerAddress string
	topic           string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&producerAddress, "producer-address", "", "")
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
	producer, err := nsq.NewProducer(producerAddress, config)
	if err != nil {
		logrus.WithError(err).Fatal("creating producer")
	}

	scheduler.Run(ctx, ds, producer, topic)
}
