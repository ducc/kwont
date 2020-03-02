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

func (s *server) GetPriceHistory(ctx context.Context, req *protos.GetPriceHistoryRequest) (*protos.GetPriceHistoryResponse, error) {

	return nil, nil
}

func (s *server) AddPriceHistory(ctx context.Context, req *protos.AddPriceHistoryRequest) (*protos.AddPriceHistoryResponse, error) {
	ts, err := ptypes.Timestamp(req.PriceChange.Timestamp)
	if err != nil {
		return nil, err
	}

	if err := s.db.InsertSymbolPrice(ctx, req.PriceChange.Symbol.String(), ts, req.PriceChange.Price); err != nil {
		return nil, err
	}

	return &protos.AddPriceHistoryResponse{}, nil

}
