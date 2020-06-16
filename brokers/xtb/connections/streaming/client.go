package streaming

import (
	"context"
	"encoding/json"
	"github.com/ducc/kwɒnt/brokers/xtb/connections"
	"github.com/sirupsen/logrus"
	"time"
)

type Client struct {
	conn            *connections.Connection
	streamSessionID string
	log             *logrus.Entry

	getTickPricesResponses  chan *GetTickPricesResponse
	getTradesResponses      chan *GetTradesResponse
	getTradeStatusResponses chan *GetTradeStatusResponse
}

func New(ctx context.Context, streamSessionID string) (*Client, error) {
	const endpoint = "wss://ws.xapi.pro/demoStream"
	conn, err := connections.New(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:                    conn,
		streamSessionID:         streamSessionID,
		log:                     logrus.WithField("client", "stream"),
		getTickPricesResponses:  make(chan *GetTickPricesResponse),
		getTradesResponses:      make(chan *GetTradesResponse),
		getTradeStatusResponses: make(chan *GetTradeStatusResponse),
	}

	go c.PingLoop()
	go c.ProcessMessages()

	// subscribe to updates on the status of new trades as they go from pending to executed
	if err := c.SendGetTradeStatus(ctx); err != nil {
		return nil, err
	}

	// subscribe to updates on new/existing trades
	if err := c.SendGetTrades(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) GetTickPricesResponses() <-chan *GetTickPricesResponse {
	return c.getTickPricesResponses
}

func (c *Client) GetTradesResponses() <-chan *GetTradesResponse {
	return c.getTradesResponses
}

func (c *Client) GetTradeStatusResponses() <-chan *GetTradeStatusResponse {
	return c.getTradeStatusResponses
}

// todo gross make a message bus or something
func (c *Client) ProcessMessages() {
	defer close(c.getTickPricesResponses)
	defer close(c.getTradesResponses)
	defer close(c.getTradeStatusResponses)

	for data := range c.conn.ReadMessages() {
		var dataMap map[string]interface{}
		if err := json.Unmarshal(data, &dataMap); err != nil {
			c.log.WithError(err).Error("unmarshalling data as map")
			continue
		}

		command, ok := dataMap["command"]
		if !ok {
			c.log.Debug("message does not have a command")
			continue
		}

		switch command {
		case "tickPrices":
			{
				var tickPrices GetTickPricesResponse
				if err := json.Unmarshal(data, &tickPrices); err != nil {
					c.log.WithError(err).Error("unmarshalling data as GetTickPricesResponse")
					continue
				}
				c.getTickPricesResponses <- &tickPrices
			}
		case "trade":
			{
				c.log.Debug("got trade msg")

				var trades GetTradesResponse
				if err := json.Unmarshal(data, &trades); err != nil {
					c.log.WithError(err).Error("unmarshalling data as GetTradesResponse")
					continue
				}
				c.getTradesResponses <- &trades
			}
		case "tradeStatus":
			{
				var tradeStatus GetTradeStatusResponse
				if err := json.Unmarshal(data, &tradeStatus); err != nil {
					c.log.WithError(err).Error("unmarshalling data as GetTradeStatusResponse")
					continue
				}
				c.getTradeStatusResponses <- &tradeStatus
			}
		default:
			c.log.WithField("command", command).Debug("unhandled command")
		}
	}
}

func (c *Client) SendPing(ctx context.Context) error {
	c.log.Debug("sending ping")

	return c.conn.WriteJSON(ctx, &PingRequest{
		Command:         "ping",
		StreamSessionID: c.streamSessionID,
	})
}

func (c *Client) PingLoop() {
	for range time.NewTicker(time.Second * 3).C {
		ctx := context.Background()

		err := c.SendPing(ctx)
		if err == connections.ErrConnectionClosed {
			c.log.Debug("connection is closed, breaking ping loop")
			break
		}
		if err != nil {
			c.log.WithError(err).Error("sending ping")
		}
	}
}

func (c *Client) SendGetTickPrices(ctx context.Context, symbol string) error {
	return c.conn.WriteJSON(ctx, &GetTickPricesRequest{
		Command:         "getTickPrices",
		StreamSessionID: c.streamSessionID,
		Symbol:          symbol,
		MinArrivalTime:  200,
		MaxLevel:        0,
	})
}

func (c *Client) SendGetNews(ctx context.Context) error {
	return c.conn.WriteJSON(ctx, &GetNewsRequest{
		Command:         "getNews",
		StreamSessionID: c.streamSessionID,
	})
}

func (c *Client) SendGetTradeStatus(ctx context.Context) error {
	return c.conn.WriteJSON(ctx, &GetTradeStatusRequest{
		Command:         "getTradeStatus",
		StreamSessionID: c.streamSessionID,
	})
}

func (c *Client) SendGetTrades(ctx context.Context) error {
	return c.conn.WriteJSON(ctx, &GetTradesRequest{
		Command:         "getTrades",
		StreamSessionID: c.streamSessionID,
	})
}
