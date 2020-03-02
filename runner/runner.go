package runner

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

type runner struct {
	ds       protos.DataServiceClient
	se       protos.StrategyEvaluatorClient
	broker   protos.BrokerServiceClient
	consumer *nsq.Consumer
	topic    string
}

func Run(ctx context.Context, ds protos.DataServiceClient, se protos.StrategyEvaluatorClient, broker protos.BrokerServiceClient, consumer *nsq.Consumer, topic string) {
	r := &runner{
		ds:       ds,
		se:       se,
		broker:   broker,
		consumer: consumer,
		topic:    topic,
	}

	consumer.AddHandler(r)
}

func (r *runner) HandleMessage(msg *nsq.Message) error {
	ctx := context.Background()

	var strategy *protos.Strategy
	if err := proto.Unmarshal(msg.Body, strategy); err != nil {
		logrus.WithError(err).Error("unmarshalling message to strategy")
		msg.Finish()
		return nil
	}

	r.getPriceHistory(ctx, strategy, msg.Finish)
	return nil
}

func (r *runner) getPriceHistory(ctx context.Context, strategy *protos.Strategy, ack func()) {
	history, err := r.ds.GetPriceHistory(ctx, &protos.GetPriceHistoryRequest{
		Symbol: strategy.Symbol,
	})
	if err != nil {
		logrus.WithError(err).Error("getting price history")
		return
	}

	if len(history.Candlesticks) == 0 {
		ack()
		return
	}

	r.evaluateStrategy(ctx, strategy, history.Candlesticks, ack)
}

func (r *runner) evaluateStrategy(ctx context.Context, strategy *protos.Strategy, history []*protos.Candlestick, ack func()) {
	res, err := r.se.Evaluate(ctx, &protos.EvaulateStrategyRequest{
		Strategy:     strategy,
		Candlesticks: history,
	})
	if err != nil {
		logrus.WithError(err).Error("evaluating strategy rules")
		return
	}

	if openPosition := res.Action.GetOpenPosition(); openPosition != nil {
		r.openPosition(ctx, strategy, openPosition, ack)
		return
	}

	if closePosition := res.Action.GetClosePosition(); closePosition != nil {
		r.closePosition(ctx, strategy, closePosition, ack)
		return
	}
}

func (r *runner) openPosition(ctx context.Context, strategy *protos.Strategy, openPosition *protos.EvaluateStrategyResponse_Action_OpenPosition, ack func()) {
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

	ack()
}

func (r *runner) closePosition(ctx context.Context, strategy *protos.Strategy, closePosition *protos.EvaluateStrategyResponse_Action_ClosePosition, ack func()) {
	index, openPosition, err := findOpenPosition(strategy)
	if err != nil {
		logrus.WithError(err).Error("finding open position")
		return
	}

	if openPosition == nil {
		logrus.Error("strategy does not have an open position")
		ack()
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

	ack()
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