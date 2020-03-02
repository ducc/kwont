package xtb

import (
	"context"
	"github.com/ducc/kwɒnt/protos"
)

type server struct {
}

func (s *server) OpenPosition(ctx context.Context, req *protos.OpenPositionRequest) (*protos.OpenPositionResponse, error) {
	return nil, nil
}

func (s *server) ClosePosition(ctx context.Context, req *protos.ClosePositionRequest) (*protos.ClosePositionResponse, error) {
	return nil, nil
}

func (s *server) GetPriceHistory(ctx context.Context, req *protos.GetPriceHistoryRequest) (*protos.GetPriceHistoryResponse, error) {
	return nil, nil
}
