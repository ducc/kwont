package xtb

import (
	"context"
	"github.com/ducc/kwɒnt/brokers/xtb/sessions"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"strconv"
	"sync"
	"time"
)

type server struct {
	protos.BrokerServiceServer
	sessionsLock sync.Mutex
	sessions     map[string]*sessions.SessionController
	tickChan     *amqp.Channel
	tradeChan    *amqp.Channel
	tickQueue    amqp.Queue
	tradeQueue   amqp.Queue
}

func New(tickChan, tradeChan *amqp.Channel, tickQueue, tradeQueue amqp.Queue, router protos.BrokerServiceClient) *server {
	s := &server{
		sessions:   make(map[string]*sessions.SessionController),
		tickChan:   tickChan,
		tradeChan:  tradeChan,
		tickQueue:  tickQueue,
		tradeQueue: tradeQueue,
	}
	go s.registerWithRouter(router)

	return s
}

func (s *server) registerWithRouter(router protos.BrokerServiceClient) {
	for range time.NewTicker(time.Second).C {
		ctx := context.Background()

		if _, err := router.RegisterBroker(ctx, &protos.RegisterBrokerRequest{
			Address: os.Getenv("POD_IP") + ":8080",
		}); err != nil {
			logrus.WithError(err).Error("registering broker with router")
		}
	}
}

func (s *server) GetCurrentSessions(ctx context.Context, req *protos.GetCurrentSessionsRequest) (*protos.GetCurrentSessionsResponse, error) {
	return &protos.GetCurrentSessionsResponse{
		SessionId: s.listSessionIDs(),
	}, nil
}

func (s *server) OpenSession(ctx context.Context, req *protos.OpenSessionRequest) (*protos.OpenSessionResponse, error) {
	session, err := s.createSession(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &protos.OpenSessionResponse{
		SessionId: session.SessionID,
	}, nil
}

var ErrSessionDoesNotExist = status.Error(codes.NotFound, "session does not exist")

func (s *server) OpenPosition(ctx context.Context, req *protos.OpenPositionRequest) (*protos.OpenPositionResponse, error) {
	session := s.getSession(req.SessionId)
	if session == nil {
		return nil, ErrSessionDoesNotExist
	}

	if err := session.OpenTradeTransaction(ctx, req.Symbol, req.Direction, req.Price, req.Voliume, req.Id); err != nil {
		return nil, err
	}

	return &protos.OpenPositionResponse{}, nil
}

func (s *server) ClosePosition(ctx context.Context, req *protos.ClosePositionRequest) (*protos.ClosePositionResponse, error) {
	session := s.getSession(req.SessionId)
	if session == nil {
		return nil, ErrSessionDoesNotExist
	}

	order, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, err
	}

	if err := session.CloseTradeTransaction(ctx, req.Symbol, order); err != nil {
		return nil, err
	}

	return &protos.ClosePositionResponse{}, nil
}

func (s *server) GetBrokerPriceHistory(ctx context.Context, req *protos.GetBrokerPriceHistoryRequest) (*protos.GetBrokerPriceHistoryResponse, error) {
	session := s.getSession(req.SessionId)
	if session == nil {
		return nil, ErrSessionDoesNotExist
	}

	// todo

	return nil, nil
}

func (s *server) SubscribeToPriceChanges(ctx context.Context, req *protos.SubscribeToPriceChangesRequest) (*protos.SubscribeToPriceChangesResponse, error) {
	session := s.getSession(req.SessionId)
	if session == nil {
		return nil, ErrSessionDoesNotExist
	}

	if err := session.AddTickSubscription(ctx, req.Symbol); err != nil {
		return nil, err
	}

	return &protos.SubscribeToPriceChangesResponse{}, nil
}

func (s *server) getSession(sessionID string) *sessions.SessionController {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()
	return s.sessions[sessionID]
}

func (s *server) listSessionIDs() []string {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()
	out := make([]string, 0)
	for k := range s.sessions {
		out = append(out, k)
	}
	return out
}

func (s *server) createSession(ctx context.Context, username, password string) (*sessions.SessionController, error) {
	session, err := sessions.New(ctx, s.tickChan, s.tradeChan, s.tickQueue, s.tradeQueue, username, password)
	if err != nil {
		return nil, err
	}

	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	s.sessions[session.SessionID] = session
	return session, nil
}
