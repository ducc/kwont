package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/brokers/xtb/connections/streaming"
	"github.com/ducc/kwɒnt/brokers/xtb/connections/transactional"
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

	txClient, err := transactional.New(ctx)
	if err != nil {
		panic(err)
	}

	go txClient.ProcessMessages()
	go txClient.PingLoop()
	Must(txClient.SendLogin(ctx, username, password))

	streamSessionID, err := txClient.WaitForStreamSessionID(ctx, time.Minute)
	if err != nil {
		panic(err)
	}

	streamClient, err := streaming.New(ctx, streamSessionID)
	if err != nil {
		panic(err)
	}

	Must(streamClient.SendGetNews(ctx))
	Must(streamClient.SendGetTickPrices(ctx, "EURUSD"))
	Must(streamClient.SendGetTickPrices(ctx, "GBPUSD"))

	/*socketClient, err := xtb.NewAPIClient(ctx, "wss://ws.xapi.pro/demo", username, password)
	if err != nil {
		panic(err)
	}

	Must(socketClient.Connect(ctx))
	go func() {
		Must(socketClient.ReadMessages())
	}()
	Must(socketClient.Login())
	go func() {
		Must(socketClient.SocketPingLoop())
	}()
	// Must(client.GetTickPrices("EURUSD"))

	go func() {
		for range time.NewTicker(time.Millisecond).C {
			if socketClient.GetState() == xtb.Ready {
				break
			}
		}

		streamClient, err := xtb.NewAPIClient(ctx, "wss://ws.xapi.pro/demoStream", username, password)
		if err != nil {
			panic(err)
		}

		streamClient.SetStreamSessionID(socketClient.GetStreamSessionID())
		Must(streamClient.Connect(ctx))
		go func() {
			Must(streamClient.ReadMessages())
		}()
		go func() {
			Must(streamClient.StreamPingLoop())
		}()
		time.Sleep(time.Second * 10)
		Must(streamClient.StreamGetTickPrices("EURUSD"))
	}()*/

	<-make(chan struct{})
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
