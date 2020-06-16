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

func TradeToProto(sessionID string, trade *streaming.GetTradesResponse) (*protos.XTBTrade, error) {
	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return nil, err
	}

	closeTime, err := ptypes.TimestampProto(time.Unix(0, trade.Data.CloseTime*1000000))
	if err != nil {
		return nil, err
	}

	expiration, err := ptypes.TimestampProto(time.Unix(0, trade.Data.Expiration*1000000))
	if err != nil {
		return nil, err
	}

	openTime, err := ptypes.TimestampProto(time.Unix(0, trade.Data.OpenTime*1000000))
	if err != nil {
		return nil, err
	}

	symbolName := ProtoFromSymbol(trade.Data.Symbol)
	if symbolName == protos.Symbol_UNKNOWN {
		return nil, ErrUnsupportedSymbol
	}

	return &protos.XTBTrade{
		Timestamp:     timestamp,
		SessionId:     sessionID,
		Order:         trade.Data.Order,
		ClosePrice:    trade.Data.ClosePrice,
		CloseTime:     closeTime,
		Closed:        trade.Data.Closed,
		Cmd:           trade.Data.Cmd.String(),
		Comment:       trade.Data.Comment,
		Commission:    trade.Data.Commission,
		CustomComment: trade.Data.CustomComment,
		Digits:        trade.Data.Digits,
		Expiration:    expiration,
		MarginRate:    trade.Data.MarginRate,
		Offset:        trade.Data.Offset,
		OpenPrice:     trade.Data.OpenPrice,
		OpenTime:      openTime,
		Order2:        trade.Data.Order2,
		Position:      trade.Data.Position,
		Profit:        trade.Data.Profit,
		StopLoss:      trade.Data.StopLoss,
		State:         trade.Data.State.String(),
		Storage:       trade.Data.Storage,
		Symbol:        symbolName,
		TakeProfit:    trade.Data.TakeProfit,
		Type:          trade.Data.Type.String(),
		Volume:        trade.Data.Volume,
	}, nil
}
