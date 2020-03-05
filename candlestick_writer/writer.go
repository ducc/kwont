package candlestick_writer

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type writer struct {
	ds protos.DataServiceClient
}

func Run(ctx context.Context, ds protos.DataServiceClient, subscription *nats.Subscription) {
	w := &writer{
		ds: ds,
	}

	w.processMessages(ctx, subscription)
}

func (w *writer) processMessages(ctx context.Context, subscription *nats.Subscription) {
	for {
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

		w.processMessage(ctx, msg)
	}
}

func (w *writer) processMessage(ctx context.Context, msg *nats.Msg) {
	var candlestick protos.Candlestick
	if err := proto.Unmarshal(msg.Data, &candlestick); err != nil {
		logrus.WithError(err).Error("unmarshalling message to candlestick")
		return
	}

	w.sendToDatabase(ctx, &candlestick)
}

func (w *writer) sendToDatabase(ctx context.Context, candlestick *protos.Candlestick) {
	logrus.WithField("candlestick", candlestick).Debug("sending candlestick to database")

	if _, err := w.ds.AddCandlestick(ctx, &protos.AddCandlestickRequest{
		Candlestick: candlestick,
	}); err != nil {
		logrus.WithError(err).Error("sending candlestick to database")
		return
	}
}
