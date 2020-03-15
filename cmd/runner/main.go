package main

import (
	"context"
	"flag"
	"github.com/ducc/kwɒnt/dataservice"
	"github.com/ducc/kwɒnt/protos"
	"github.com/ducc/kwɒnt/runner"
	"github.com/golang/protobuf/ptypes"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

var (
	level                    string
	natsAddress              string
	natsUsername             string
	natsPassword             string
	topic                    string
	strategyEvaluatorAddress string
)

func init() {
	flag.StringVar(&level, "level", "debug", "")
	flag.StringVar(&natsAddress, "nats-address", "127.0.0.1:4150", "nats server address")
	flag.StringVar(&natsUsername, "nats-username", "kwont", "nats username")
	flag.StringVar(&natsPassword, "nats-password", "password", "nats password")
	flag.StringVar(&topic, "topic", "", "")
	flag.StringVar(&strategyEvaluatorAddress, "strategy-evaluator-address", "127.0.0.1:50051", "address of strategy evaluator service")
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

	se, err := newStrategyEvaluatorClient(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("creating strategy evaluator client")
	}

	natsConn, err := nats.Connect(natsAddress, nats.UserInfo(natsUsername, natsPassword))
	if err != nil {
		logrus.WithError(err).Fatal("connecting to nats")
	}

	subscription, err := natsConn.SubscribeSync(topic)
	if err != nil {
		logrus.WithError(err).Fatal("subscribing to topic")
	}

	broker := &MockBrokerServiceClient{}

	runner.Run(ctx, ds, se, broker, subscription, topic)
}

func newStrategyEvaluatorClient(ctx context.Context) (protos.StrategyEvaluatorClient, error) {
	logrus.WithField("address", strategyEvaluatorAddress).Debug("connecting to strategy evaluator")
	conn, err := grpc.DialContext(ctx, strategyEvaluatorAddress, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	logrus.Debug("connected to strategy evaluator")
	return protos.NewStrategyEvaluatorClient(conn), nil
}

type MockBrokerServiceClient struct {
	protos.BrokerServiceClient
}

func (c *MockBrokerServiceClient) OpenPosition(ctx context.Context, in *protos.OpenPositionRequest, opts ...grpc.CallOption) (*protos.OpenPositionResponse, error) {
	ts, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}

	return &protos.OpenPositionResponse{
		Id:             "1234",
		ExecutionTime:  ts,
		ExecutionPrice: 123,
	}, nil
}

func (c *MockBrokerServiceClient) ClosePosition(ctx context.Context, in *protos.ClosePositionRequest, opts ...grpc.CallOption) (*protos.ClosePositionResponse, error) {
	ts, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}

	return &protos.ClosePositionResponse{
		ExecutionTime:  ts,
		ExecutionPrice: 456,
	}, nil
}
