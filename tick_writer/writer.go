package tick_writer

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type writer struct {
	ds protos.DataServiceClient
}

func Run(ctx context.Context, ds protos.DataServiceClient, messages <-chan amqp.Delivery) {
	w := &writer{
		ds: ds,
	}

	w.processMessages(ctx, messages)
}

func (w *writer) processMessages(ctx context.Context, messages <-chan amqp.Delivery) {
	for msg := range messages {
		if msg.Body == nil || len(msg.Body) == 0 {
			logrus.Debug("body is nil or empty")
			continue
		}

		w.processMessage(ctx, msg)
	}
}

func (w *writer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var tick protos.Tick
	if err := proto.Unmarshal(msg.Body, &tick); err != nil {
		logrus.WithError(err).Error("unmarshalling message to tick")
		return
	}

	if err := w.sendToDatabase(ctx, &tick); err != nil {
		logrus.WithError(err).Error("sending tick to database")
		return
	}

	if err := msg.Ack(false); err != nil {
		logrus.WithError(err).Error("acking message")
	}
}

func (w *writer) sendToDatabase(ctx context.Context, tick *protos.Tick) error {
	logrus.WithField("tick", tick).Debug("sending tick to database")

	if _, err := w.ds.AddTick(ctx, &protos.AddTickRequest{
		Tick: tick,
	}); err != nil {
		return err
	}

	return nil
}
