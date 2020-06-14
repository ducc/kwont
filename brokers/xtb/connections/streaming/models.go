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

type GetTradesRequest struct {
	Command         string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
}

type GetTradesResponse struct {
	Command string                      `json:"command"`
	Data    *GetTradeStatusResponseData `json:"data"`
}

type GetTradesResponseData struct {
	ClosePrice    float64                  `json:"close_price"`
	CloseTime     int64                    `json:"close_time"`
	Closed        bool                     `json:"closed"`
	Cmd           GetTradesResponseDataCmd `json:"cmd"`
	Comment       string                   `json:"comment"`
	Commission    float64                  `json:"commission"`
	CustomComment string                   `json:"customComment"`
	Digits        int64                    `json:"digits"`
	Expiration    int64                    `json:"expiration"`
	MarginRate    float64                  `json:"margin_rate"`
	Offset        int64                    `json:"offset"`
	OpenPrice     float64                  `json:"open_price"`
	OpenTime      int64                    `json:"open_time"`
	Order         int64                    `json:"order"`
	Order2        int64                    `json:"order2"`
	Position      int64                    `json:"position"`
	Profit        float64                  `json:"profit"`
	StopLoss      float64                  `json:"sl"`
	State         GetTradesResponseDataCmd
	Storage       float64                   `json:"storage"`
	Symbol        string                    `json:"symbol"`
	TakeProfit    float64                   `json:"tp"`
	Type          GetTradesResponseDataType `json:"type"`
	Volume        float64                   `json:"volume"`
}

type GetTradesResponseDataCmd int

const (
	GetTradesResponseDataCmd_BUY        GetTradesResponseDataCmd = 0
	GetTradesResponseDataCmd_SELL       GetTradesResponseDataCmd = 1
	GetTradesResponseDataCmd_BUY_LIMIT  GetTradesResponseDataCmd = 2
	GetTradesResponseDataCmd_SELL_LIMIT GetTradesResponseDataCmd = 3
	GetTradesResponseDataCmd_BUY_STOP   GetTradesResponseDataCmd = 4
	GetTradesResponseDataCmd_SELL_STOP  GetTradesResponseDataCmd = 5
	GetTradesResponseDataCmd_BALANCE    GetTradesResponseDataCmd = 6
	GetTradesResponseDataCmd_CREDIT     GetTradesResponseDataCmd = 7
)

type GetTradesResponseDataState string

const (
	GetTradesResponseDataState_MODIFIED GetTradesResponseDataState = "Modified"
	GetTradesResponseDataState_DELETED  GetTradesResponseDataState = "Deleted"
)

type GetTradesResponseDataType int

const (
	GetTradesResponseDataType_OPEN    GetTradesResponseDataType = 0
	GetTradesResponseDataType_PENDING GetTradesResponseDataType = 1
	GetTradesResponseDataType_CLOSE   GetTradesResponseDataType = 2
	GetTradesResponseDataType_MODIFY  GetTradesResponseDataType = 3
	GetTradesResponseDataType_DELETE  GetTradesResponseDataType = 4
)
