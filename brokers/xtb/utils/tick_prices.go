package utils

import (
	"github.com/ducc/kwɒnt/brokers/xtb/connections/streaming"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/ptypes"
	"time"
)

func TickPriceToProto(tickPrice *streaming.GetTickPricesResponse) (*protos.Tick, error) {
	// convert timestamp to nanoseconds
	timestampTime := time.Unix(0, tickPrice.Data.Timestamp*1000000)
	timestamp, err := ptypes.TimestampProto(timestampTime)
	if err != nil {
		return nil, err
	}

	symbolName := ProtoFromSymbol(tickPrice.Data.Symbol)
	if symbolName == protos.Symbol_UNKNOWN {
		return nil, ErrUnsupportedSymbol
	}

	return &protos.Tick{
		Timestamp:  timestamp,
		Broker:     protos.Broker_XTB_DEMO,
		Symbol:     symbolName,
		Price:      tickPrice.Data.Ask,
		Spread:     tickPrice.Data.SpreadRaw,
		BuyVolume:  float64(tickPrice.Data.AskVolume),
		SellVolume: float64(tickPrice.Data.BidVolume),
	}, nil
}
