package router

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type router struct {
	protos.BrokerServiceServer
	finder      *SessionFinder
	connections *BrokerConnections
}

func NewRouter() *router {
	connections := NewBrokerConnections()

	r := &router{
		finder:      NewSessionFinder(),
		connections: connections,
	}

	go r.pollSessions()

	return r
}

func (r *router) pollSessions() {
	for range time.NewTicker(time.Second * 3).C {
		ctx := context.Background()

		addressesToRemove := make([]string, 0)

		for address, client := range r.connections.GetConnections() {
			res, err := client.GetCurrentSessions(ctx, &protos.GetCurrentSessionsRequest{})
			if err != nil {
				// if err means client no longer exits remove it and make sure the sessions have been remapped
				addressesToRemove = append(addressesToRemove, address)
				logrus.WithError(err).Error("getting current sessions")
				continue
			}

			r.connections.SetActiveSessions(address, int64(len(res.SessionId)))

			for _, sessionID := range res.SessionId {
				r.finder.SetServiceAddress(sessionID, address)
			}
		}

		for _, address := range addressesToRemove {
			r.connections.RemoveConnection(address)

			sessions := r.finder.GetSessionsForAddress(address)
			for _, sessionID := range sessions {
				r.finder.RemoveSession(sessionID)
			}
		}
	}
}

func (r *router) RegisterBroker(ctx context.Context, req *protos.RegisterBrokerRequest) (*protos.RegisterBrokerResponse, error) {
	if _, err := r.connections.GetOrConnect(ctx, req.Address); err != nil {
		return nil, err
	}

	return &protos.RegisterBrokerResponse{}, nil
}

func (r *router) getConnection(ctx context.Context, sessionID string) (protos.BrokerServiceClient, error) {
	serviceAddress := r.finder.GetServiceAddress(sessionID)
	if serviceAddress == "" {
		return nil, status.Error(codes.NotFound, "session does not exist")
	}

	return r.connections.GetOrConnect(ctx, serviceAddress)
}

func (r *router) GetCurrentSessions(ctx context.Context, req *protos.GetCurrentSessionsRequest) (*protos.GetCurrentSessionsResponse, error) {
	sessionIDs := r.finder.GetSessionIds()
	return &protos.GetCurrentSessionsResponse{
		SessionId: sessionIDs,
	}, nil
}

func (r *router) OpenSession(ctx context.Context, req *protos.OpenSessionRequest) (*protos.OpenSessionResponse, error) {
	serviceAddress := r.connections.FindAddressWithLeastSessions()
	conn, err := r.connections.GetOrConnect(ctx, serviceAddress)
	if err != nil {
		return nil, err
	}

	res, err := conn.OpenSession(ctx, req)
	if err != nil {
		return nil, err
	}

	r.finder.SetServiceAddress(res.SessionId, serviceAddress)
	return res, nil
}

func (r *router) OpenPosition(ctx context.Context, req *protos.OpenPositionRequest) (*protos.OpenPositionResponse, error) {
	client, err := r.getConnection(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}

	return client.OpenPosition(ctx, req)
}

func (r *router) ClosePosition(ctx context.Context, req *protos.ClosePositionRequest) (*protos.ClosePositionResponse, error) {
	client, err := r.getConnection(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}

	return client.ClosePosition(ctx, req)
}

func (r *router) GetBrokerPriceHistory(ctx context.Context, req *protos.GetBrokerPriceHistoryRequest) (*protos.GetBrokerPriceHistoryResponse, error) {
	client, err := r.getConnection(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return client.GetBrokerPriceHistory(ctx, req)
}

func (r *router) SubscribeToPriceChanges(ctx context.Context, req *protos.SubscribeToPriceChangesRequest) (*protos.SubscribeToPriceChangesResponse, error) {
	client, err := r.getConnection(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}
	return client.SubscribeToPriceChanges(ctx, req)
}
