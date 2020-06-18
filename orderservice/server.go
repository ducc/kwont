package orderservice

import (
	"context"
	"errors"
	"fmt"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
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

func (s *server) getUserSessionID(ctx context.Context, userID string, broker protos.Broker_Name) (string, error) {
	// todo change this to just get the user's broker session id as we know the user and the broker
	usersResponse, err := s.data.ListUsers(ctx, &protos.ListUsersRequest{})
	if err != nil {
		return "", err
	}
	var user *protos.User
	for _, u := range usersResponse.Users {
		if u.Id == userID {
			user = u
			break
		}
	}
	if user == nil {
		return "", errors.New("invalid user")
	}

	var sessionID string
	for _, conn := range user.BrokerConnections {
		if conn.Broker == broker {
			sessionID = conn.SessionId
			break
		}
	}
	if sessionID == "" {
		// todo should a session be created now as it doesn't currently exist?
		return "", errors.New("user does not have an open session")
	}

	return sessionID, nil
}

func (s *server) OpenPosition(ctx context.Context, req *protos.OpenPositionRequest) (*protos.OpenPositionResponse, error) {
	ts, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}

	dsResponse, err := s.data.AddOrder(ctx, &protos.AddOrderRequest{
		Order: &protos.Order{
			Broker:    req.Broker,
			Symbol:    req.Symbol,
			Direction: req.Direction,
			Price:     req.Price,
			Volume:    req.Voliume,
			Timestamp: ts,
		},
	})
	if err != nil {
		return nil, err
	}

	sessionID, err := s.getUserSessionID(ctx, req.UserId, req.Broker)
	if err != nil {
		return nil, err
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
	order, err := s.data.GetOrder(ctx, &protos.GetOrderRequest{
		OrderId: req.Id,
	})
	if err != nil {
		return nil, err
	}

	res, err := s.data.GetXTBTrades(ctx, &protos.GetXTBTradesRequest{
		OrderId: req.Id,
	})
	if err != nil {
		return nil, err
	}

	if len(res.Trades) == 0 {
		return nil, status.Error(codes.NotFound, "order does not exist in xtb_trades table")
	}

	var trade *protos.XTBTrade
	for _, t := range res.Trades {
		if t.Type == "OPEN" && t.State == "MODIFIED" {
			trade = t
		}
	}

	if trade == nil {
		return nil, status.Error(codes.NotFound, "no trade with open type and modified state")
	}

	sessionID, err := s.getUserSessionID(ctx, req.UserId, req.Broker)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("closing order %d", trade.Order)
	return s.broker.ClosePosition(ctx, &protos.ClosePositionRequest{
		SessionId: sessionID,
		Symbol:    order.Order.Symbol,
		Id:        fmt.Sprint(trade.Order),
		Direction: order.Order.Direction,
		Price:     order.Order.Price,
		Volume:    order.Order.Volume,
	})
}
