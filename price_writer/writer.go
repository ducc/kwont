package price_writer

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
)

type writer struct {
	ds       protos.DataServiceClient
	consumer *nsq.Consumer
	topic    string
}

func Run(ctx context.Context, ds protos.DataServiceClient, consumer *nsq.Consumer, topic string) {
	w := &writer{
		ds:       ds,
		consumer: consumer,
		topic:    topic,
	}

	consumer.AddHandler(w)
}

func (w *writer) HandleMessage(msg *nsq.Message) error {
	ctx := context.Background()

	var priceChange *protos.PriceChange
	if err := proto.Unmarshal(msg.Body, priceChange); err != nil {
		logrus.WithError(err).Error("unmarshalling message to price change")
		msg.Finish()
		return nil
	}

	w.sendToDatabase(ctx, priceChange, msg.Finish)
	return nil
}

func (w *writer) sendToDatabase(ctx context.Context, priceChange *protos.PriceChange, ack func()) {
	_, err := w.ds.AddPriceHistory(ctx, &protos.AddPriceHistoryRequest{
		PriceChange: priceChange,
	})
	if err != nil {
		logrus.WithError(err).Error("sending price history to database")
		return
	}

	ack()
}
