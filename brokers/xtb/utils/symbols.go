package utils

import "github.com/ducc/kwɒnt/protos"

var protoToSymbol = map[protos.Symbol_Name]string{
	protos.Symbol_EUR_USD: "EURUSD",
	protos.Symbol_USD_GBP: "GBPUSD",
}

func SymbolFromProto(symbol protos.Symbol_Name) string {
	return protoToSymbol[symbol]
}

func ProtoFromSymbol(symbol string) protos.Symbol_Name {
	for k, v := range protoToSymbol {
		if v == symbol {
			return k
		}
	}
	return protos.Symbol_UNKNOWN
}
