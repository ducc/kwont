package router

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"time"
)

type router struct {
	protos.BrokerServiceClient
	broker      protos.Broker_Name
	finder      *SessionFinder
	connections *BrokerConnections
}

func NewRouter(finder *SessionFinder, broker protos.Broker_Name) *router {
	connections := NewBrokerConnections()

	r := &router{
		broker:      broker,
		finder:      finder,
		connections: connections,
	}

	go r.pollSessions()

	return r
}

func (r *router) pollSessions() {
	for range time.NewTicker(time.Second * 3).C {
		ctx := context.Background()

		for address, client := range r.connections.GetConnections() {
			res, err := client.GetCurrentSessions(ctx, &protos.GetCurrentSessionsRequest{})
			if err != nil {
				// todo if err means client no longer exits remove it and make sure the sessions have been remapped
				logrus.WithError(err).Error("getting current sessions")
				continue
			}

			for _, sessionID := range res.SessionId {
				_, err := r.finder.getSessionInfo(ctx, r.broker, sessionID)
				if err == ErrSessionNotFound {
					if err := r.finder.setSessionInfo(ctx, &protos.SessionInfo{
						SessionId:      sessionID,
						Broker:         r.broker,
						ServiceAddress: address,
					}); err != nil {
						logrus.WithError(err).Error("adding session to redis")
						continue
					}
					continue
				}
				if err != nil {
					logrus.WithError(err).Error("getting session info")
					continue
				}
			}
		}
	}
}

func (r *router) OpenSession(ctx context.Context, req *protos.OpenSessionRequest) (*protos.OpenSessionResponse, error) {
	// todo talk to load balancer
	return nil, nil
}

func (r *router) OpenPosition(ctx context.Context, req *protos.OpenPositionRequest) (*protos.OpenPositionResponse, error) {
	sessionInfo, err := r.finder.getSessionInfo(ctx, req.Symbol.Broker, req.SessionId)
	if err != nil {
		return nil, err
	}

	client, err := r.connections.GetOrConnect(ctx, sessionInfo.ServiceAddress)
	if err != nil {
		return nil, err
	}

	return client.OpenPosition(ctx, req)
}

func (r *router) ClosePosition(ctx context.Context, req *protos.ClosePositionRequest) (*protos.ClosePositionResponse, error) {
	sessionInfo, err := r.finder.getSessionInfo(ctx, req.Broker, req.SessionId)
	if err != nil {
		return nil, err
	}

	client, err := r.connections.GetOrConnect(ctx, sessionInfo.ServiceAddress)
	if err != nil {
		return nil, err
	}

	return client.ClosePosition(ctx, req)
}

func (r *router) GetBrokerPriceHistory(ctx context.Context, req *protos.GetBrokerPriceHistoryRequest) (*protos.GetBrokerPriceHistoryResponse, error) {
	sessionInfo, err := r.finder.getSessionInfo(ctx, req.Symbol.Broker, req.SessionId)
	if err != nil {
		return nil, err
	}

	client, err := r.connections.GetOrConnect(ctx, sessionInfo.ServiceAddress)
	if err != nil {
		return nil, err
	}

	return client.GetBrokerPriceHistory(ctx, req)
}
