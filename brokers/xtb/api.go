package xtb

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

type connectionState int

const (
	Disconnected connectionState = iota
	AwaitingStreamSessionID
	Ready
)

type apiClient struct {
	endpoint, username, password string
	conn                         *websocket.Conn
	state                        connectionState
	streamSessionID              string
	sync.Mutex
}

func NewAPIClient(ctx context.Context, endpoint, username, password string) (*apiClient, error) {
	return &apiClient{
		endpoint: endpoint,
		username: username,
		password: password,
	}, nil
}

func (c *apiClient) Connect(ctx context.Context) error {
	logrus.Debug("connecting")
	conn, res, err := websocket.DefaultDialer.DialContext(ctx, c.endpoint, nil)
	if err != nil {
		return err
	}

	logrus.Debug(*res)
	c.conn = conn
	return nil
}

type loginMessage struct {
	Command   string                 `json:"command"`
	Arguments *LoginMessageArguments `json:"arguments"`
}

type LoginMessageArguments struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
	AppID    string `json:"appId"`
	AppName  string `json:"appName"`
}

func (c *apiClient) Login() error {
	c.Lock()
	defer c.Unlock()

	logrus.Debug("logging in")
	msg := &loginMessage{
		Command: "login",
		Arguments: &LoginMessageArguments{
			UserID:   c.username,
			Password: c.password,
			AppName:  "kwont",
		},
	}

	c.state = AwaitingStreamSessionID

	if err := c.conn.WriteJSON(msg); err != nil {
		return err
	}

	return nil
}

func (c *apiClient) ReadMessages() error {
	defer func() {
		if err := c.conn.Close(); err != nil {
			logrus.WithError(err).Error("closing websocket connection")
		}
	}()

	logrus.Debug("reading messages")
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}

		message := string(data)
		logrus.Debug(message)

		if err := c.HandleMessage(data); err != nil {
			return err
		}
	}
}

func (c *apiClient) HandleMessage(data []byte) error {
	c.Lock()
	defer c.Unlock()

	switch c.state {
	case AwaitingStreamSessionID:
		if strings.Contains(string(data), "streamSessionId") {
			var loginResponse LoginResponse
			if err := json.Unmarshal(data, &loginResponse); err != nil {
				return err
			}
			c.streamSessionID = loginResponse.StreamSessionID
			c.state = Ready
		}
	}

	return nil
}

type PingMessage struct {
	Command         string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
}

func (c *apiClient) PingLoop() error {
	for range time.NewTicker(time.Second * 10).C {
		c.Lock()
		if c.state != Ready {
			c.Unlock()
			continue
		}

		if err := c.Ping(); err != nil {
			c.Unlock()
			return err
		}
		c.Unlock()
	}

	return nil
}

func (c *apiClient) Ping() error {
	logrus.Debug("pinging")

	return c.conn.WriteJSON(&PingMessage{
		Command:         "ping",
		StreamSessionID: c.streamSessionID,
	})
}

type LoginResponse struct {
	Status          bool   `json:"status"`
	StreamSessionID string `json:"streamSessionId"`
}

type getTickPricesMessage struct {
	Command   string                 `json:"command"`
	Arguments getTickPricesArguments `json:"arguments"`
}

type getTickPricesArguments struct {
	Level     int64    `json:"level"`
	Symbols   []string `json:"symbols"`
	Timestamp int64    `json:"timestamp"`
}

func (c *apiClient) GetTickPrices() error {
	logrus.Debug("getting tick prices")

	return c.conn.WriteJSON(&getTickPricesMessage{
		Command: "getTickPrices",
		Arguments: getTickPricesArguments{
			Level:     0,
			Symbols:   []string{"EURUSD"},
			Timestamp: 1583012235550,
		},
	})
}
