package sessions

import (
	"context"
	"github.com/ducc/kwɒnt/brokers/xtb/connections/streaming"
	"github.com/ducc/kwɒnt/brokers/xtb/connections/transactional"
	"github.com/ducc/kwɒnt/brokers/xtb/utils"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type Session struct {
	sync.Mutex

	tickChan  *amqp.Channel
	tickQueue amqp.Queue

	tradeChan  *amqp.Channel
	tradeQueue amqp.Queue

	SessionID string
	username  string
	password  string

	tx     *transactional.Client
	stream *streaming.Client

	finished  chan struct{}
	startTime time.Time

	tickSubscriptions map[protos.Symbol_Name]bool
}

func newSession(ctx context.Context, tickChan, tradeChan *amqp.Channel, tickQueue, tradeQueue amqp.Queue, username, password, sessionID string) (*Session, error) {
	tx, err := transactional.New(ctx)
	if err != nil {
		panic(err)
	}

	go tx.ProcessMessages()
	go tx.PingLoop()

	if err := tx.SendLogin(ctx, username, password); err != nil {
		return nil, err
	}

	streamSessionID, err := tx.WaitForStreamSessionID(ctx, time.Minute)
	if err != nil {
		return nil, err
	}

	stream, err := streaming.New(ctx, streamSessionID)
	if err != nil {
		return nil, err
	}

	s := &Session{
		tickChan:          tickChan,
		tickQueue:         tickQueue,
		tradeChan:         tradeChan,
		tradeQueue:        tradeQueue,
		SessionID:         sessionID,
		username:          username,
		password:          password,
		tx:                tx,
		stream:            stream,
		finished:          make(chan struct{}, 1),
		startTime:         time.Now(),
		tickSubscriptions: make(map[protos.Symbol_Name]bool),
	}

	go s.transformTickPricesToProto()

	for symbolIndex := range protos.Symbol_Name_name {
		symbol := protos.Symbol_Name(symbolIndex)
		if symbol == protos.Symbol_UNKNOWN {
			continue
		}

		if err := s.AddTickSubscription(ctx, symbol); err != nil && err != utils.ErrUnsupportedSymbol {
			return nil, err
		}
	}

	return s, nil
}

func (s *Session) AddTickSubscription(ctx context.Context, symbol protos.Symbol_Name) error {
	symbolName := utils.SymbolFromProto(symbol)
	if symbolName == "" {
		return utils.ErrUnsupportedSymbol
	}

	if err := s.stream.SendGetTickPrices(ctx, symbolName); err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()
	s.tickSubscriptions[symbol] = true

	return nil
}

func (s *Session) OpenTradeTransaction(ctx context.Context, symbol protos.Symbol_Name, direction protos.Direction_Name, price, volume float64, orderID string) error {
	symbolName := utils.SymbolFromProto(symbol)
	if symbolName == "" {
		return utils.ErrUnsupportedSymbol
	}

	return s.tx.OpenTradeTransaction(ctx, symbolName, direction, price, volume, orderID)
}

func (s *Session) CloseTradeTransaction(ctx context.Context, symbol protos.Symbol_Name, order int64) error {
	symbolName := utils.SymbolFromProto(symbol)
	if symbolName == "" {
		return utils.ErrUnsupportedSymbol
	}

	return s.tx.CloseTradeTransaction(ctx, symbolName, order)
}

func (s *Session) GetTickSubscription() []protos.Symbol_Name {
	s.Lock()
	defer s.Unlock()

	copied := make([]protos.Symbol_Name, 0, len(s.tickSubscriptions))
	for symbolName := range s.tickSubscriptions {
		copied = append(copied, symbolName)
	}

	return copied
}

func (s *Session) transformTickPricesToProto() {
	defer func() {
		s.finished <- struct{}{}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for tickPrice := range s.stream.GetTickPricesResponses() {
			ctx := context.Background()

			tick, err := utils.TickPriceToProto(tickPrice)
			if err != nil {
				logrus.WithError(err).Error("converting tick price to proto")
				continue
			}

			s.sendTickToQueue(ctx, tick)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for trade := range s.stream.GetTradesResponses() {
			logrus.Debug("processing trade response")
			ctx := context.Background()

			trade, err := utils.TradeToProto(s.SessionID, trade)
			if err != nil {
				logrus.WithError(err).Error("converting trade to proto")
				continue
			}

			s.sendTradeToQueue(ctx, trade)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for tradeStatus := range s.stream.GetTradeStatusResponses() {
			_ = tradeStatus
		}
	}()

	wg.Wait()
}

func (s *Session) sendTickToQueue(ctx context.Context, tick *protos.Tick) {
	bytes, err := proto.Marshal(tick)
	if err != nil {
		logrus.WithError(err).Error("error marshalling tick")
		return
	}

	if err := s.tickChan.Publish(
		"",               // exchange
		s.tickQueue.Name, // routing key
		false,            // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         bytes,
		}); err != nil {
		logrus.WithError(err).Error("publishing tick amqp message")
	}
}

func (s *Session) sendTradeToQueue(ctx context.Context, trade *protos.XTBTrade) {
	logrus.Debug("sending trade to queue")

	bytes, err := proto.Marshal(trade)
	if err != nil {
		logrus.WithError(err).Error("error marshalling trade")
		return
	}

	if err := s.tradeChan.Publish(
		"",                // exchange
		s.tradeQueue.Name, // routing key
		false,             // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         bytes,
		}); err != nil {
		logrus.WithError(err).Error("publishing trade amqp message")
	}
}
