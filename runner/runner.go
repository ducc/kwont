package runner

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type runner struct {
	ds           protos.DataServiceClient
	se           protos.StrategyEvaluatorClient
	broker       protos.BrokerServiceClient
	subscription *nats.Subscription
	topic        string
}

func Run(ctx context.Context, ds protos.DataServiceClient, se protos.StrategyEvaluatorClient, broker protos.BrokerServiceClient, subscription *nats.Subscription, topic string) {
	r := &runner{
		ds:           ds,
		se:           se,
		broker:       broker,
		subscription: subscription,
		topic:        topic,
	}

	r.processMessages(subscription)
}

func (r *runner) processMessages(subscription *nats.Subscription) {
	for {
		ctx := context.Background()

		msg, err := subscription.NextMsgWithContext(ctx)
		if err != nil {
			logrus.WithError(err).Fatal("getting next message")
		}

		if msg == nil {
			logrus.Debug("received nil message")
			continue
		}

		if msg.Data == nil || len(msg.Data) == 0 {
			logrus.Debug("data is nil or empty")
			continue
		}

		r.processMessage(ctx, msg)
	}
}

func (r *runner) processMessage(ctx context.Context, msg *nats.Msg) {
	var strategy *protos.Strategy
	if err := proto.Unmarshal(msg.Data, strategy); err != nil {
		logrus.WithError(err).Error("unmarshalling message to strategy")
		return
	}

	r.getPriceHistory(ctx, strategy)
}

func (r *runner) getPriceHistory(ctx context.Context, strategy *protos.Strategy) {
	history, err := r.ds.GetPriceHistory(ctx, &protos.GetPriceHistoryRequest{
		Symbol: strategy.Symbol,
		// todo start, end
		WindowNanoseconds: 300000000000, // todo find lowest window from all rules so it can be rewindowed in the strategy evaluator server
	})
	if err != nil {
		logrus.WithError(err).Error("getting price history")
		return
	}

	if len(history.Candlesticks) == 0 {
		logrus.Debug("get price history returned no candlesticks")
		return
	}

	r.evaluateStrategy(ctx, strategy, history.Candlesticks)
}

func (r *runner) evaluateStrategy(ctx context.Context, strategy *protos.Strategy, history []*protos.Candlestick) {
	res, err := r.se.Evaluate(ctx, &protos.EvaulateStrategyRequest{
		Strategy:     strategy,
		Candlesticks: history,
	})
	if err != nil {
		logrus.WithError(err).Error("evaluating strategy rules")
		return
	}

	if openPosition := res.Action.GetOpenPosition(); openPosition != nil {
		r.openPosition(ctx, strategy, openPosition)
		return
	}

	if closePosition := res.Action.GetClosePosition(); closePosition != nil {
		r.closePosition(ctx, strategy, closePosition)
		return
	}
}

func (r *runner) openPosition(ctx context.Context, strategy *protos.Strategy, openPosition *protos.EvaluateStrategyResponse_Action_OpenPosition) {
	res, err := r.broker.OpenPosition(ctx, &protos.OpenPositionRequest{
		Direction: openPosition.Direction,
		Price:     openPosition.Price,
	})
	if err != nil {
		logrus.WithError(err).Error("opening position")
		return
	}

	strategy.Positions = append(strategy.Positions, &protos.Position{
		Direction: openPosition.Direction,
		OpenPrice: res.ExecutionPrice,
		OpenTime:  res.ExecutionTime,
		Id:        res.Id,
	})

	if _, err := r.ds.UpdateStrategy(ctx, &protos.UpdateStrategyRequest{
		Strategy: strategy,
	}); err != nil {
		logrus.WithError(err).Error("updating strategy")
		return
	}
}

func (r *runner) closePosition(ctx context.Context, strategy *protos.Strategy, closePosition *protos.EvaluateStrategyResponse_Action_ClosePosition) {
	index, openPosition, err := findOpenPosition(strategy)
	if err != nil {
		logrus.WithError(err).Error("finding open position")
		return
	}

	if openPosition == nil {
		logrus.Error("strategy does not have an open position")
		return
	}

	res, err := r.broker.ClosePosition(ctx, &protos.ClosePositionRequest{
		Id:    openPosition.Id,
		Price: closePosition.Price,
	})
	if err != nil {
		logrus.WithError(err).Error("opening position")
		return
	}

	pos := strategy.Positions[index]
	pos.CloseTime = res.ExecutionTime
	pos.ClosePrice = res.ExecutionPrice

	if _, err := r.ds.UpdateStrategy(ctx, &protos.UpdateStrategyRequest{
		Strategy: strategy,
	}); err != nil {
		logrus.WithError(err).Error("updating strategy")
		return
	}
}

func findOpenPosition(strategy *protos.Strategy) (int, *protos.Position, error) {
	for i, position := range strategy.Positions {
		closeTime, err := ptypes.Timestamp(position.CloseTime)
		if err != nil {
			return i, nil, err
		}

		if !closeTime.IsZero() {
			return i, position, nil
		}
	}

	return -1, nil, nil
}
