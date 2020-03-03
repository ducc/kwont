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

	getTickPricesResponses chan *GetTickPricesResponse
}

func New(ctx context.Context, streamSessionID string) (*Client, error) {
	const endpoint = "wss://ws.xapi.pro/demoStream"
	conn, err := connections.New(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:                   conn,
		streamSessionID:        streamSessionID,
		log:                    logrus.WithField("client", "stream"),
		getTickPricesResponses: make(chan *GetTickPricesResponse),
	}

	go c.PingLoop()
	go c.ProcessMessages()

	return c, nil
}

func (c *Client) GetTickPricesResponses() <-chan *GetTickPricesResponse {
	return c.getTickPricesResponses
}

func (c *Client) ProcessMessages() {
	defer close(c.getTickPricesResponses)

	for data := range c.conn.ReadMessages() {
		var dataMap map[string]interface{}
		if err := json.Unmarshal(data, dataMap); err != nil {
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
		MinArrivalTime:  5000,
		MaxLevel:        0,
	})
}

func (c *Client) SendGetNews(ctx context.Context) error {
	return c.conn.WriteJSON(ctx, &GetNewsRequest{
		Command:         "getNews",
		StreamSessionID: c.streamSessionID,
	})
}
