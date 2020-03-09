package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()

	c, err := dataservice.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	res, err := c.GetPriceHistory(ctx, &protos.GetPriceHistoryRequest{
		Symbol: &protos.Symbol{
			Broker: protos.Broker_XTB_DEMO,
			Name:   protos.Symbol_BITCOIN,
		},
		WindowNanoseconds: int64(time.Hour),
	})
	if err != nil {
		panic(err)
	}

	file, err := os.Create("candlesticks.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	defer buf.Flush()

	w := csv.NewWriter(buf)
	defer w.Flush()

	w.Write([]string{"time", "low", "high", "open", "close"})

	for _, candlestick := range res.Candlesticks {
		ts, err := ptypes.Timestamp(candlestick.Timestamp)
		if err != nil {
			panic(err)
		}

		w.Write([]string{ts.Format(time.Stamp), fmt.Sprint(candlestick.Low), fmt.Sprint(candlestick.High), fmt.Sprint(candlestick.Open), fmt.Sprint(candlestick.Close)})
	}
}
