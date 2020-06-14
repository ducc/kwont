package transactional

type LoginRequest struct {
	Command   string                 `json:"command"`
	Arguments *LoginRequestArguments `json:"arguments"`
}

type LoginRequestArguments struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
	AppID    string `json:"appId"`
	AppName  string `json:"appName"`
}

type LoginResponse struct {
	Status          bool   `json:"status"`
	StreamSessionID string `json:"streamSessionId"`
}

type PingRequest struct {
	Command string `json:"command"`
}

type TradeTransactionRequest struct {
	Command   string                     `json:"command"`
	Arguments *TradeTransactionArguments `json:"arguments"`
}

type TradeTransactionArguments struct {
	TradeTransInfo *TradeTransactionInfo `json:"tradeTransInfo"`
}

type TradeTransactionInfo struct {
	Cmd           TradeTransactionInfoOperationCode `json:"cmd,omitempty"`
	CustomComment string                            `json:"customCommand,omitempty"`
	Expiration    int64                             `json:"expiration,omitempty"`
	Offset        int64                             `json:"offset,omitempty"`
	Order         int64                             `json:"order"`
	Price         float64                           `json:"price,omitempty"`
	StopLoss      float64                           `json:"sl,omitempty"`
	Symbol        string                            `json:"symbol"`
	TakeProfit    float64                           `json:"tp,omitempty"`
	Type          TradeTransactionInfoType          `json:"type"`
	Volume        float64                           `json:"volume,omitempty"`
}

type TradeTransactionInfoOperationCode int

const (
	TradeTransactionInfoOperationCode_BUY        = 0
	TradeTransactionInfoOperationCode_SELL       = 1
	TradeTransactionInfoOperationCode_BUY_LIMIT  = 2
	TradeTransactionInfoOperationCode_SELL_LIMIT = 3
	TradeTransactionInfoOperationCode_BUY_STOP   = 4
	TradeTransactionInfoOperationCode_SELL_STOP  = 5
	TradeTransactionInfoOperationCode_BALANCE    = 6
	TradeTransactionInfoOperationCode_CREDIT     = 7
)

type TradeTransactionInfoType int

const (
	TradeTransactionInfoType_OPEN    = 0
	TradeTransactionInfoType_PENDING = 1
	TradeTransactionInfoType_CLOSE   = 2
	TradeTransactionInfoType_MODIFY  = 3
	TradeTransactionInfoType_DELETE  = 4
)

type TradeTransactionResponse struct {
	Status     bool                       `json:"status"`
	ReturnData TradeTransactionReturnData `json:"returnData"`
}

type TradeTransactionReturnData struct {
	Order int `json:"order"`
}
