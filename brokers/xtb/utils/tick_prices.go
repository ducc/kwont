package utils

import (
	"github.com/ducc/kwɒnt/brokers/xtb/connections/streaming"
	"github.com/ducc/kwɒnt/protos"
	"github.com/golang/protobuf/ptypes"
	"time"
)

func TickPriceToCandlestick(tickPrice *streaming.GetTickPricesResponse) (*protos.Candlestick, error) {
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

	return &protos.Candlestick{
		Timestamp: timestamp,
		Symbol: &protos.Symbol{
			Name:   symbolName,
			Broker: protos.Broker_XTB_DEMO,
		},
		Current:    PoundsToMicros(tickPrice.Data.Ask),
		Spread:     PoundsToMicros(tickPrice.Data.SpreadRaw),
		BuyVolume:  int64(tickPrice.Data.AskVolume),
		SellVolume: int64(tickPrice.Data.BidVolume),
	}, nil
}
