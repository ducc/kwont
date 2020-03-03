package xtb

import (
	"context"
	"github.com/ducc/kwɒnt/brokers/xtb/sessions"
	"github.com/ducc/kwɒnt/protos"
	"github.com/nsqio/go-nsq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
)

type server struct {
	sessionsLock sync.Mutex
	sessions     map[string]*sessions.Session
	producer     *nsq.Producer
}

func New(producer *nsq.Producer) *server {
	return &server{
		sessions: make(map[string]*sessions.Session),
		producer: producer,
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

	// todo

	return nil, nil
}

func (s *server) ClosePosition(ctx context.Context, req *protos.ClosePositionRequest) (*protos.ClosePositionResponse, error) {
	session := s.getSession(req.SessionId)
	if session == nil {
		return nil, ErrSessionDoesNotExist
	}

	// todo

	return nil, nil
}

func (s *server) GetBrokerPriceHistory(ctx context.Context, req *protos.GetBrokerPriceHistoryRequest) (*protos.GetBrokerPriceHistoryResponse, error) {
	session := s.getSession(req.SessionId)
	if session == nil {
		return nil, ErrSessionDoesNotExist
	}

	// todo

	return nil, nil
}

func (s *server) getSession(sessionID string) *sessions.Session {
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

func (s *server) createSession(ctx context.Context, username, password string) (*sessions.Session, error) {
	session, err := sessions.New(ctx, username, password)
	if err != nil {
		return nil, err
	}

	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()

	s.sessions[session.SessionID] = session
	return session, nil
}
