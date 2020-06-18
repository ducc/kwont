package trade_writer

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/ducc/kwɒnt/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

type writer struct {
	ds     protos.DataServiceClient
	broker protos.Broker_Name
}

func Run(ctx context.Context, ds protos.DataServiceClient, broker protos.Broker_Name, messages <-chan *pubsub.Message) {
	w := &writer{
		ds:     ds,
		broker: broker,
	}

	w.processMessages(ctx, messages)
}

func (w *writer) processMessages(ctx context.Context, messages <-chan *pubsub.Message) {
	for msg := range messages {
		if msg.Body() == nil || len(msg.Body()) == 0 {
			logrus.Debug("body is nil or empty")
			continue
		}

		w.processMessage(ctx, msg)
	}
}

func (w *writer) processMessage(ctx context.Context, msg *pubsub.Message) {
	switch w.broker {
	case protos.Broker_XTB_DEMO:
		var xtbTrade protos.XTBTrade
		if err := proto.Unmarshal(msg.Body(), &xtbTrade); err != nil {
			logrus.WithError(err).Error("unmarshalling message to xtb trade")
			return
		}

		if err := w.sendXTBTradeToDatabase(ctx, &xtbTrade); err != nil {
			logrus.WithError(err).Error("sending xtb trade to database")
			return
		}
	default:
		logrus.Fatal("unsupported broker")
	}

	if err := msg.Ack(); err != nil {
		logrus.WithError(err).Error("acking message")
	}
}

func (w *writer) sendXTBTradeToDatabase(ctx context.Context, trade *protos.XTBTrade) error {
	logrus.WithField("trade", trade).Debug("sending xtb trade to database")

	if _, err := w.ds.AddXTBTrade(ctx, &protos.AddXTBTradeRequest{
		Trade: trade,
	}); err != nil {
		return err
	}

	return nil
}
