package utils

import "github.com/ducc/kwɒnt/protos"

var protoToSymbol = map[protos.Symbol_Name]string{
	protos.Symbol_AUD_USD: "AUDUSD",
	protos.Symbol_EUR_USD: "EURUSD",
	protos.Symbol_EUR_CHF: "EURCHF",
	protos.Symbol_GBP_JPY: "GBPJPY",
	protos.Symbol_USD_CHF: "USDCHF",
	protos.Symbol_USD_GBP: "USDGBP",
	protos.Symbol_USD_CAD: "USDCAD",
	protos.Symbol_USD_JPY: "USDJPY",

	protos.Symbol_BITCOIN:  "BITCOIN",
	protos.Symbol_LITECOIN: "LITECOIN",
	protos.Symbol_ETHEREUM: "ETHEREUM",

	protos.Symbol_US_30:  "US30",
	protos.Symbol_UK_100: "UK100",
	protos.Symbol_DE_30:  "DE30",

	protos.Symbol_SILVER: "SILVER",
	protos.Symbol_GOLD:   "GOLD",
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
