package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	flag.Parse()

	ds, err := dataservice.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	se, err := newStrategyEvaluatorClient(ctx)
	if err != nil {
		panic(err)
	}

	strategies, err := ds.ListStrategies(ctx, &protos.ListStrategiesRequest{})
	if err != nil {
		panic(err)
	}

	strat := strategies.Strategies[0]
	logrus.Debug(strat)

	candlesticks, err := ds.GetPriceHistory(ctx, &protos.GetPriceHistoryRequest{
		Symbol: &protos.Symbol{
			Broker: protos.Broker_XTB_DEMO,
			Name:   protos.Symbol_BITCOIN,
		},
		WindowNanoseconds: int64(time.Minute * 5),
	})
	if err != nil {
		panic(err)
	}

	logrus.Debugf("candlesticks %d", len(candlesticks.Candlesticks))

	res, err := se.Evaluate(ctx, &protos.EvaulateStrategyRequest{
		Strategy:        strat,
		Candlesticks:    candlesticks.Candlesticks,
		HasOpenPosition: false,
	})
	if err != nil {
		panic(err)
	}

	logrus.Debug(res.Action.GetOpenPosition())
	logrus.Debug(res.Action.GetClosePosition())
}

func newStrategyEvaluatorClient(ctx context.Context) (protos.StrategyEvaluatorClient, error) {
	logrus.Debug("connecting to strategy evaluator")
	conn, err := grpc.DialContext(ctx, "127.0.0.1:50051", grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	logrus.Debug("connected to strategy evaluator")
	return protos.NewStrategyEvaluatorClient(conn), nil

}
