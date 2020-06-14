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

type GetTradeStatusRequest struct {
	Command         string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
}

type GetTradeStatusResponse struct {
	Command string                      `json:"command"`
	Data    *GetTradeStatusResponseData `json:"data"`
}

type GetTradeStatusResponseData struct {
	CustomComment string                      `json:"customComment"`
	Message       string                      `json:"message"`
	Order         int                         `json:"order"`
	Price         float64                     `json:"price"`
	RequestStatus GetTradeStatusRequestStatus `json:"requestStatus"`
}

type GetTradeStatusRequestStatus int

const (
	GetTradeStatusRequestStatus_ERROR    = 0
	GetTradeStatusRequestStatus_PENDING  = 1
	GetTradeStatusRequestStatus_ACCEPTED = 3
	GetTradeStatusRequestStatus_REJECTED = 4
)
