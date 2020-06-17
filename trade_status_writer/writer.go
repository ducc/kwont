package trade_status_writer

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type writer struct {
	ds     protos.DataServiceClient
	broker protos.Broker_Name
}

func Run(ctx context.Context, ds protos.DataServiceClient, broker protos.Broker_Name, messages <-chan amqp.Delivery) {
	w := &writer{
		ds:     ds,
		broker: broker,
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
	switch w.broker {
	case protos.Broker_XTB_DEMO:
		var xtbTradeStatus protos.XTBTradeStatus
		if err := proto.Unmarshal(msg.Body, &xtbTradeStatus); err != nil {
			logrus.WithError(err).Error("unmarshalling message to xtb trade")
			return
		}

		if err := w.sendXTBTradeStatusToDatabase(ctx, &xtbTradeStatus); err != nil {
			logrus.WithError(err).Error("sending xtb trade status to database")
			return
		}
	default:
		logrus.Fatal("unsupported broker")
	}

	if err := msg.Ack(false); err != nil {
		logrus.WithError(err).Error("acking message")
	}
}

func (w *writer) sendXTBTradeStatusToDatabase(ctx context.Context, status *protos.XTBTradeStatus) error {
	logrus.WithField("status", status).Debug("sending xtb trade status to database")

	if _, err := w.ds.AddXTBTradeStatus(ctx, &protos.AddXTBTradeStatusRequest{
		Status: status,
	}); err != nil {
		return err
	}

	return nil
}
