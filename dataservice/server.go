package dataservice

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
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

/*func (s *server) CreateStrategy(ctx context.Context, req *protos.CreateStrategyRequest) (*protos.CreateStrategyResponse, error) {
	entryRulesBytes, err := proto.Marshal(req.Strategy.EntryRules)
	if err != nil {
		return nil, err
	}

	exitRulesBytes, err := proto.Marshal(req.Strategy.ExitRules)
	if err != nil {
		return nil, err
	}

	strategyID, err := s.db.InsertStrategy(ctx, entryRulesBytes, exitRulesBytes, req.Strategy.Status.String(), req.Strategy.Name, req.Strategy.Symbol.String())
	if err != nil {
		return nil, err
	}

	return &protos.CreateStrategyResponse{
		Id: strategyID,
	}, nil
}

func (s *server) UpdateStrategy(ctx context.Context, req *protos.UpdateStrategyRequest) (*protos.UpdateStrategyResponse, error) {

	return nil, nil
}

func (s *server) ListStrategies(ctx context.Context, req *protos.ListStrategiesRequest) (*protos.ListStrategiesResponse, error) {

	return nil, nil
}

*/
func (s *server) GetPriceHistory(ctx context.Context, req *protos.GetPriceHistoryRequest) (*protos.GetPriceHistoryResponse, error) {
	partials, err := s.db.GetPartialCandlesticks(ctx, req.Symbol.Name.String(), req.Symbol.Broker.String(), time.Now().Add((time.Hour*12)*-1), time.Now())
	if err != nil {
		return nil, err
	}

	windows := make(map[time.Time][]*protos.Candlestick)

	for _, partial := range partials {
		timestamp, err := ptypes.Timestamp(partial.Timestamp)
		if err != nil {
			return nil, err
		}

		windowTime := timestamp.Truncate(time.Duration(req.WindowNanoseconds))
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
		jTimestamp, err = ptypes.Timestamp(aggregated[i].Timestamp)

		return iTimestamp.Before(jTimestamp)
	})

	return &protos.GetPriceHistoryResponse{
		Candlesticks: aggregated,
	}, nil
}

func (s *server) AddCandlestick(ctx context.Context, req *protos.AddCandlestickRequest) (*protos.AddCandlestickResponse, error) {
	c := req.Candlestick

	ts, err := ptypes.Timestamp(c.Timestamp)
	if err != nil {
		return nil, err
	}

	if err := s.db.InsertCandlestick(ctx, c.Symbol.Name.String(), c.Symbol.Broker.String(), ts, c.Open, c.Close, c.High, c.Low, c.Current, c.Spread, c.BuyVolume, c.SellVolume); err != nil {
		return nil, err
	}

	return &protos.AddCandlestickResponse{}, nil
}
