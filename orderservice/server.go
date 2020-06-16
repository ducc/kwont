package orderservice

import (
	"context"
	"errors"
	"github.com/ducc/kwɒnt/protos"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	protos.OrderServiceServer
	data   protos.DataServiceClient
	broker protos.BrokerServiceClient
}

func NewServer(ctx context.Context, data protos.DataServiceClient, broker protos.BrokerServiceClient) (*server, error) {
	return &server{
		data:   data,
		broker: broker,
	}, nil
}

func (s *server) OpenPosition(ctx context.Context, req *protos.OpenPositionRequest) (*protos.OpenPositionResponse, error) {
	dsResponse, err := s.data.AddOrder(ctx, &protos.AddOrderRequest{
		Broker:    req.Broker,
		Symbol:    req.Symbol,
		Direction: req.Direction,
		Price:     req.Price,
		Volume:    req.Voliume,
	})
	if err != nil {
		return nil, err
	}

	// todo change this to just get the user's broker session id as we know the user and the broker
	usersResponse, err := s.data.ListUsers(ctx, &protos.ListUsersRequest{})
	if err != nil {
		return nil, err
	}
	var user *protos.User
	for _, u := range usersResponse.Users {
		if u.Id == req.UserId {
			user = u
			break
		}
	}
	if user == nil {
		return nil, errors.New("invalid user")
	}

	var sessionID string
	for _, conn := range user.BrokerConnections {
		if conn.Broker == req.Broker {
			sessionID = conn.SessionId
			break
		}
	}
	if sessionID == "" {
		// todo should a session be created now as it doesn't currently exist?
		return nil, errors.New("user does not have an open session")
	}

	// todo it is bad that both services use the same proto message
	if _, err := s.broker.OpenPosition(ctx, &protos.OpenPositionRequest{
		SessionId: sessionID,
		Id:        dsResponse.OrderId,
		Direction: req.Direction,
		Symbol:    req.Symbol,
		Price:     req.Price,
		Voliume:   req.Voliume,
	}); err != nil {
		return nil, err
	}

	return &protos.OpenPositionResponse{
		Id: dsResponse.OrderId,
	}, nil
}

func (s *server) ClosePosition(ctx context.Context, req *protos.ClosePositionRequest) (*protos.ClosePositionResponse, error) {
	// todo implement position closing
	return nil, status.Error(codes.Unimplemented, "cba rn")
}
