package dataservice

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

type server struct {
	protos.DataServiceServer
	db *database
}

func NewServer(ctx context.Context, databaseAddress string) (*server, error) {
	db, err := newDatabase(ctx, databaseAddress)
	if err != nil {
		return nil, err
	}

	return &server{
		db: db,
	}, nil
}

func (s *server) CreateStrategy(ctx context.Context, req *protos.CreateStrategyRequest) (*protos.CreateStrategyResponse, error) {
	entryRulesBytes, err := proto.Marshal(req.Strategy.EntryRules)
	if err != nil {
		return nil, err
	}

	exitRulesBytes, err := proto.Marshal(req.Strategy.ExitRules)
	if err != nil {
		return nil, err
	}

	strategyID, err := s.db.InsertStrategy(ctx, entryRulesBytes, exitRulesBytes, req.Strategy.Status.String(), req.Strategy.Name, req.Strategy.Symbol.Name.String(), req.Strategy.Symbol.Broker.String())
	if err != nil {
		return nil, err
	}

	return &protos.CreateStrategyResponse{
		Id: strategyID,
	}, nil
}

func (s *server) UpdateStrategy(ctx context.Context, req *protos.UpdateStrategyRequest) (*protos.UpdateStrategyResponse, error) {
	entryRulesBytes, err := proto.Marshal(req.Strategy.EntryRules)
	if err != nil {
		return nil, err
	}

	exitRulesBytes, err := proto.Marshal(req.Strategy.ExitRules)
	if err != nil {
		return nil, err
	}

	ts, err := ptypes.Timestamp(req.Strategy.LastEvaluated)
	if err != nil {
		return nil, err
	}

	if err := s.db.UpdateStrategy(ctx, req.Strategy.Id, entryRulesBytes, exitRulesBytes, req.Strategy.Status.String(), req.Strategy.Name, req.Strategy.Symbol.Name.String(), req.Strategy.Symbol.Broker.String(), ts); err != nil {
		return nil, err
	}

	return &protos.UpdateStrategyResponse{}, nil
}

func (s *server) ListStrategies(ctx context.Context, req *protos.ListStrategiesRequest) (*protos.ListStrategiesResponse, error) {
	strategies, err := s.db.ListStrategies(ctx)
	if err != nil {
		return nil, err
	}

	return &protos.ListStrategiesResponse{
		Strategies: strategies,
	}, nil
}

func (s *server) GetPriceHistory(ctx context.Context, req *protos.GetPriceHistoryRequest) (*protos.GetPriceHistoryResponse, error) {
	start, err := ptypes.Timestamp(req.Start)
	if err != nil {
		return nil, err
	}

	end, err := ptypes.Timestamp(req.End)
	if err != nil {
		return nil, err
	}

	candlesticks, err := s.db.GetCandlesticks(ctx, req.Window, req.Broker, req.Symbol, start, end)
	if err != nil {
		return nil, err
	}

	return &protos.GetPriceHistoryResponse{
		Candlesticks: candlesticks,
	}, nil
}

func (s *server) AddTick(ctx context.Context, req *protos.AddTickRequest) (*protos.AddTickResponse, error) {
	t := req.Tick

	ts, err := ptypes.Timestamp(t.Timestamp)
	if err != nil {
		return nil, err
	}

	if err := s.db.InsertTick(ctx, ts, t.Broker.String(), t.Symbol.String(), t.Price, t.Spread, t.BuyVolume, t.SellVolume); err != nil {
		return nil, err
	}

	if err := s.db.InsertOrUpdateCandlestick(ctx, protos.CandlestickWindow_ONE_MINUTE, ts, t.Broker.String(), t.Symbol.String(), t.Price, t.Spread, t.BuyVolume, t.SellVolume); err != nil {
		return nil, err
	}

	if err := s.db.InsertOrUpdateCandlestick(ctx, protos.CandlestickWindow_ONE_HOUR, ts, t.Broker.String(), t.Symbol.String(), t.Price, t.Spread, t.BuyVolume, t.SellVolume); err != nil {
		return nil, err
	}

	if err := s.db.InsertOrUpdateCandlestick(ctx, protos.CandlestickWindow_ONE_DAY, ts, t.Broker.String(), t.Symbol.String(), t.Price, t.Spread, t.BuyVolume, t.SellVolume); err != nil {
		return nil, err
	}

	return &protos.AddTickResponse{}, nil
}

func (s *server) CreateUser(ctx context.Context, req *protos.CreateUserRequest) (*protos.CreateUserResponse, error) {
	userID, err := s.db.InsertUser(ctx, req.User.Name)
	if err != nil {
		return nil, err
	}

	return &protos.CreateUserResponse{
		Id: userID,
	}, nil
}

func (s *server) UpdateUser(ctx context.Context, req *protos.UpdateUserRequest) (*protos.UpdateUserResponse, error) {
	newUser := req.User
	userID := newUser.Id

	oldUser, err := s.db.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if oldUser.Name != newUser.Name {
		if err := s.db.UpdateUser(ctx, userID, newUser.Name); err != nil {
			return nil, err
		}
	}

	oldConnections := make(map[protos.Broker_Name]*protos.User_BrokerConnection)
	for _, oldConnection := range oldUser.BrokerConnections {
		oldConnections[oldConnection.Broker] = oldConnection
	}

	for _, newConnection := range newUser.BrokerConnections {
		// todo deleting connections

		var create bool
		var update bool

		oldConnection := oldConnections[newConnection.Broker]
		if oldConnection == nil {
			create = true
		} else {
			if oldConnection.SessionId != newConnection.SessionId {
				update = true
			} else if oldConnection.Username != newConnection.Username {
				update = true
			} else if oldConnection.Password != newConnection.Password {
				update = true
			}
		}

		if create {
			if err := s.db.InsertBrokerConnections(ctx, userID, newConnection.Broker.String(), newConnection.Username, newConnection.Password); err != nil {
				return nil, err
			}
		} else if update {
			if err := s.db.UpdateBrokerConnection(ctx, userID, newConnection.Broker.String(), newConnection.Username, newConnection.Password, newConnection.SessionId); err != nil {
				return nil, err
			}
		}
	}

	return &protos.UpdateUserResponse{}, nil
}

func (s *server) ListUsers(ctx context.Context, req *protos.ListUsersRequest) (*protos.ListUsersResponse, error) {
	users, err := s.db.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	return &protos.ListUsersResponse{
		Users: users,
	}, nil
}

func (s *server) AddOrder(ctx context.Context, req *protos.AddOrderRequest) (*protos.AddOrderResponse, error) {
	ts, err := ptypes.Timestamp(req.Timestamp)
	if err != nil {
		return nil, err
	}

	orderID, err := s.db.InsertOrder(ctx, req.Broker.String(), req.Symbol.String(), req.Direction.String(), req.Price, req.Volume, ts)
	if err != nil {
		return nil, err
	}

	return &protos.AddOrderResponse{OrderId: orderID}, nil
}

func (s *server) AddXTBTrade(ctx context.Context, req *protos.AddXTBTradeRequest) (*protos.AddXTBTradeResponse, error) {
	ts, err := ptypes.Timestamp(req.Trade.Timestamp)
	if err != nil {
		return nil, err
	}

	closeTime, err := ptypes.Timestamp(req.Trade.CloseTime)
	if err != nil {
		return nil, err
	}

	expiration, err := ptypes.Timestamp(req.Trade.Expiration)
	if err != nil {
		return nil, err
	}

	openTime, err := ptypes.Timestamp(req.Trade.OpenTime)
	if err != nil {
		return nil, err
	}

	if err := s.db.InsertXTBTrade(ctx, ts, req.Trade.SessionId, req.Trade.Order, req.Trade.ClosePrice, closeTime, req.Trade.Closed, req.Trade.Cmd, req.Trade.Comment, req.Trade.Commission, req.Trade.CustomComment, req.Trade.Digits, expiration, req.Trade.MarginRate, req.Trade.Offset, req.Trade.OpenPrice, openTime, req.Trade.Order2, req.Trade.Position, req.Trade.Profit, req.Trade.StopLoss, req.Trade.State, req.Trade.Storage, req.Trade.Symbol.String(), req.Trade.TakeProfit, req.Trade.Type, req.Trade.Volume); err != nil {
		return nil, err
	}

	return &protos.AddXTBTradeResponse{}, nil
}

func (s *server) AddXTBTradeStatus(ctx context.Context, req *protos.AddXTBTradeStatusRequest) (*protos.AddXTBTradeStatusResponse, error) {
	ts, err := ptypes.Timestamp(req.Status.Timestamp)
	if err != nil {
		return nil, err
	}

	if err := s.db.InsertXTBTradeStatus(ctx, ts, req.Status.SessionId, req.Status.Order, req.Status.CustomComment, req.Status.Message, req.Status.Price, req.Status.RequestStatus); err != nil {
		return nil, err
	}

	return &protos.AddXTBTradeStatusResponse{}, nil
}
