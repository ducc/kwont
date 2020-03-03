package streaming

type PingRequest struct {
	Command         string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
}

type GetTickPricesRequest struct {
	Command         string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
	Symbol          string `json:"symbol"`
	MinArrivalTime  int    `json:"minArrivalTime"`
	MaxLevel        int    `json:"maxLevel"`
}

type GetTickPricesResponse struct {
	Command string                    `json:"command"`
	Data    GetTickPricesResponseData `json:"data"`
}

type GetTickPricesResponseData struct {
	Symbol      string  `json:"symbol"`
	Ask         float64 `json:"ask"`
	Bid         float64 `json:"bid"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	AskVolume   int     `json:"askVolume"`
	BidVolume   int     `json:"bidVolume"`
	Timestamp   int64   `json:"timestamp"`
	Level       int     `json:"level"`
	QuoteID     int     `json:"quoteId"`
	SpreadTable float64 `json:"spreadTable"`
	SpreadRaw   float64 `json:"spreadRaw"`
}

/*
{
    "command": "tickPrices",
    "data": {
        "symbol": "EURUSD",
        "ask": 1.11885,
        "bid": 1.11883,
        "high": 1.12130,
        "low": 1.10953,
        "askVolume": 50000,
        "bidVolume": 200000,
        "timestamp": 1583263696428,
        "level": 0,
        "quoteId": 10,
        "spreadTable": 0.2,
        "spreadRaw": 0.00002
    }
}
*/

type GetNewsRequest struct {
	Command         string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
}
