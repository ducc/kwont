package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/brokers"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/ducc/kwɒnt/session_checker"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	level               string
	routerAddress       string
	brokerName          string
	pollIntervalSeconds int64
)

func init() {
	flag.StringVar(&level, "level", "debug", "logrus logging level")
	flag.StringVar(&routerAddress, "router-address", "", "router address")
	flag.StringVar(&brokerName, "broker-name", "", "")
	flag.Int64Var(&pollIntervalSeconds, "poll-interval-seconds", 15, "interval between polling for users to check")
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

	router, err := brokers.NewClient(ctx, routerAddress)
	if err != nil {
		logrus.WithError(err).Fatal("creating router client")
	}

	brokerName := protos.Broker_Name(protos.Broker_Name_value[brokerName])
	if brokerName == protos.Broker_UNKNOWN {
		logrus.Fatalf("unknown broker %s", brokerName)
	}

	session_checker.Run(ctx, ds, router, brokerName, time.Duration(pollIntervalSeconds)*time.Second)
}
