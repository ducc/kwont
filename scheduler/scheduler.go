package scheduler

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"time"
)

type scheduler struct {
	ds       protos.DataServiceClient
	natsConn *nats.Conn
	topic    string
}

func Run(ctx context.Context, ds protos.DataServiceClient, natsConn *nats.Conn, topic string, pollInterval time.Duration) {
	r := &scheduler{
		ds:       ds,
		natsConn: natsConn,
		topic:    topic,
	}

	r.pollStrategies(ctx, pollInterval)
}

func (r *scheduler) pollStrategies(ctx context.Context, pollInterval time.Duration) {
	for {
		r.findStrategyToSchedule(ctx)
		time.Sleep(pollInterval)
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
		if strategy == nil {
			logrus.Error("nil strategy")
			continue
		}

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

	r.sendStrategyForProcessing(ctx, strategy)
}

func (r *scheduler) sendStrategyForProcessing(ctx context.Context, strategy *protos.Strategy) {
	logrus.WithField("strategy", *strategy).Debug("sending strategy for processing")

	data, err := proto.Marshal(strategy)
	if err != nil {
		logrus.WithError(err).Error("marshalling strategy to proto bytes")
		return
	}

	if err := r.natsConn.Publish(r.topic, data); err != nil {
		logrus.WithError(err).Error("sending strategy to topic")
		return
	}

	lastEval, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		logrus.WithError(err).Error("converting time to proto")
		return
	}

	strategy.LastEvaluated = lastEval

	// todo this will be a race condititon
	if _, err := r.ds.UpdateStrategy(ctx, &protos.UpdateStrategyRequest{
		Strategy: strategy,
	}); err != nil {
		logrus.WithError(err).Error("updating strategy")
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

	var rulesSet *protos.RuleSet
	if hasOpenPosition {
		rulesSet = strategy.ExitRules
	} else {
		rulesSet = strategy.EntryRules
	}

	var shortest int64
	for _, rule := range rulesSet.Rules {
		if shortest == 0 || rule.PeriodNanoseconds < shortest {
			shortest = rule.PeriodNanoseconds
		}
	}

	return time.Duration(shortest), nil
}
