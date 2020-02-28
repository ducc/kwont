package scheduler

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"time"
)

type scheduler struct {
	ds       protos.DataServiceClient
	producer *nsq.Producer
	topic    string
}

func Run(ctx context.Context, ds protos.DataServiceClient, producer *nsq.Producer, topic string) {
	r := &scheduler{
		ds:       ds,
		producer: producer,
		topic:    topic,
	}

	r.pollStrategies(ctx)
}

func (r *scheduler) pollStrategies(ctx context.Context) {
	for {
		r.findStrategyToSchedule(ctx)
	}
}

func (r *scheduler) findStrategyToSchedule(ctx context.Context) {
	// todo request params
	strategies, err := r.ds.ListStrategies(ctx, &protos.ListStrategiesRequest{})
	if err != nil {
		logrus.WithError(err).Error("finding strategies to schedule")
		return
	}

	if len(strategies.Strategies) == 0 {
		return
	}

	for _, strategy := range strategies.Strategies {
		r.processStrategy(ctx, strategy)
	}
}

func (r *scheduler) processStrategy(ctx context.Context, strategy *protos.Strategy) {
	lastEvaluatedTime, err := ptypes.Timestamp(strategy.LastEvaluated)
	if err != nil {
		logrus.WithError(err).Error("parsing strategy last evaluated time")
		return
	}

	rulePeriod, err := findShortestRulePeriod(strategy)
	if err != nil {
		logrus.WithError(err).Error("finding shortest rule period")
		return
	}

	if lastEvaluatedTime.Add(rulePeriod).After(time.Now()) {
		return
	}

	// todo send strategy to evaluator pubsub queue :)
}

func (r *scheduler) sendStrategyForProcessing(ctx context.Context, strategy *protos.Strategy) {
	data, err := proto.Marshal(strategy)
	if err != nil {
		logrus.WithError(err).Error("marshalling strategy to proto bytes")
		return
	}

	if err := r.producer.Publish(r.topic, data); err != nil {
		logrus.WithError(err).Error("sending strategy to topic")
		return
	}
}

func hasOpenPosition(strategy *protos.Strategy) (bool, error) {
	for _, position := range strategy.Positions {
		closeTime, err := ptypes.Timestamp(position.CloseTime)
		if err != nil {
			return false, err
		}

		if closeTime.IsZero() {
			return true, nil
		}
	}

	return false, nil
}

func findShortestRulePeriod(strategy *protos.Strategy) (time.Duration, error) {
	hasOpenPosition, err := hasOpenPosition(strategy)
	if err != nil {
		return 0, err
	}

	var rulesSet []*protos.Rule
	if hasOpenPosition {
		rulesSet = strategy.ExitRules
	} else {
		rulesSet = strategy.EntryRules
	}

	var shortest int64
	for _, rule := range rulesSet {
		if shortest == 0 || rule.PeriodNanoseconds < shortest {
			shortest = rule.PeriodNanoseconds
		}
	}

	return time.Duration(shortest), nil
}
