package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/brokers/xtb"
	"github.com/sirupsen/logrus"
	"time"
)

var username string
var password string

func init() {
	flag.StringVar(&username, "username", "", "xopen hub username")
	flag.StringVar(&password, "password", "", "xopen hub password")
}

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)

	ctx := context.Background()
	client, err := xtb.NewAPIClient(ctx, "wss://ws.xapi.pro/demo", username, password)
	if err != nil {
		panic(err)
	}

	Must(client.Connect(ctx))
	go func() {
		Must(client.ReadMessages())
	}()
	Must(client.Login())
	go func() {
		Must(client.PingLoop())
	}()
	time.Sleep(time.Second * 10)
	Must(client.GetTickPrices())

	<-make(chan struct{})
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
