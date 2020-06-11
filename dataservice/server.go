package dataservice

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"sort"
	"time"
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
	partials, err := s.db.GetPartialCandlesticks(ctx, req.Symbol.Name.String(), req.Symbol.Broker.String(), time.Now().Add((time.Hour*12)*-1), time.Now())
	if err != nil {
		return nil, err
	}

	windowDuration := time.Duration(req.WindowNanoseconds)

	windows := make(map[time.Time][]*protos.Candlestick)

	for _, partial := range partials {
		timestamp, err := ptypes.Timestamp(partial.Timestamp)
		if err != nil {
			return nil, err
		}

		windowTime := timestamp.Truncate(windowDuration)
		window, ok := windows[windowTime]
		if !ok {
			window = make([]*protos.Candlestick, 0)
		}

		window = append(window, partial)
		windows[windowTime] = window
	}

	aggregated := make([]*protos.Candlestick, 0, len(windows))

	for windowTime, window := range windows {
		var high, low, open, close int64

		for i, partial := range window {
			if i == 0 {
				open = partial.Current
			}

			if i == len(window)-1 {
				close = partial.Current
			}

			if partial.Current > high {
				high = partial.Current
			}

			if partial.Current < low || low == 0 {
				low = partial.Current
			}
		}

		ts, err := ptypes.TimestampProto(windowTime)
		if err != nil {
			return nil, err
		}

		aggregated = append(aggregated, &protos.Candlestick{
			Timestamp: ts,
			Symbol:    req.Symbol,
			High:      high,
			Low:       low,
			Open:      open,
			Close:     close,
		})
	}

	sort.Slice(aggregated, func(i, j int) bool {
		var iTimestamp time.Time
		var jTimestamp time.Time

		iTimestamp, err = ptypes.Timestamp(aggregated[i].Timestamp)
		jTimestamp, err = ptypes.Timestamp(aggregated[j].Timestamp)

		return iTimestamp.Before(jTimestamp)
	})
	if err != nil {
		return nil, err
	}

	return &protos.GetPriceHistoryResponse{
		Candlesticks: aggregated,
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
