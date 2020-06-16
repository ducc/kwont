package transactional

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ducc/kwɒnt/brokers/xtb/connections"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const appName = "kwont"

type Client struct {
	conn *connections.Connection
	log  *logrus.Entry

	sync.RWMutex
	// empty until login response in received
	streamSessionID string
}

func New(ctx context.Context) (*Client, error) {
	const endpoint = "wss://ws.xapi.pro/demo"
	conn, err := connections.New(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
		log:  logrus.WithField("client", "tx"),
	}, nil
}

func (c *Client) ProcessMessages() {
	for data := range c.conn.ReadMessages() {
		c.processMessage(data)
	}
}

func (c *Client) processMessage(data []byte) {
	c.RLock()
	hasStreamSessionID := c.streamSessionID != ""
	c.RUnlock()

	if !hasStreamSessionID {
		if strings.Contains(string(data), "streamSessionId") {
			var loginResponse LoginResponse
			if err := json.Unmarshal(data, &loginResponse); err != nil {
				c.log.WithError(err).Error("parsing message as LoginResponse")
				return
			}

			c.Lock()
			c.streamSessionID = loginResponse.StreamSessionID
			c.Unlock()

		} else {
			// we dont care about any messages until we are logged in
			// todo timeout so we arent waiting forever
			return
		}
	}

	_ = data // todo
}

func (c *Client) WaitForStreamSessionID(ctx context.Context, timeout time.Duration) (string, error) {
	c.RLock()
	if id := c.streamSessionID; id != "" {
		c.RUnlock()
		return id, nil
	}
	c.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for range time.NewTicker(time.Millisecond * 100).C {
		c.RLock()
		if id := c.streamSessionID; id != "" {
			c.RUnlock()
			return id, nil
		}
		c.RUnlock()

		if ctx.Err() == context.Canceled {
			return "", ctx.Err()
		}

		if c.conn.Closed() {
			return "", connections.ErrConnectionClosed
		}
	}

	return "", nil // todo check this in unreachable
}

func (c *Client) SendPing(ctx context.Context) error {
	c.log.Debug("sending ping")

	return c.conn.WriteJSON(ctx, &PingRequest{
		Command: "ping",
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

func (c *Client) SendLogin(ctx context.Context, username, password string) error {
	c.log.Debug("sending login message")

	msg := &LoginRequest{
		Command: "login",
		Arguments: &LoginRequestArguments{
			UserID:   username,
			Password: password,
			AppName:  appName,
		},
	}

	return c.conn.WriteJSON(ctx, msg)
}

func (c *Client) OpenTradeTransaction(ctx context.Context, symbol string, direction protos.Direction_Name, price, volume float64) error {
	c.log.Debug("sending open trade transaction message")

	info := &TradeTransactionInfo{
		Symbol: symbol,
		Price:  price,
		Volume: volume,
		Type:   TradeTransactionInfoType_OPEN,
	}

	switch direction {
	case protos.Direction_BUY:
		info.Cmd = TradeTransactionInfoOperationCode_BUY
	case protos.Direction_SELL:
		info.Cmd = TradeTransactionInfoOperationCode_SELL
	default:
		return errors.New("unknown direction")
	}

	msg := &TradeTransactionRequest{
		Command: "tradeTransaction",
		Arguments: &TradeTransactionArguments{
			TradeTransInfo: info,
		},
	}

	return c.conn.WriteJSON(ctx, msg)
}

func (c *Client) CloseTradeTransaction(ctx context.Context, symbol string, order int64) error {
	c.log.Debug("sending close trade transaction message")

	msg := &TradeTransactionRequest{
		Command: "tradeTransaction",
		Arguments: &TradeTransactionArguments{
			TradeTransInfo: &TradeTransactionInfo{
				Order:  order,
				Type:   TradeTransactionInfoType_CLOSE,
				Symbol: symbol,
			},
		},
	}

	return c.conn.WriteJSON(ctx, msg)
}
