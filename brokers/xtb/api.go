package xtb

import (
	"context"
	"encoding/json"
	"errors"
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

func (c *apiClient) GetState() connectionState {
	c.Lock()
	defer c.Unlock()

	return c.state
}

func (c *apiClient) GetStreamSessionID() string {
	c.Lock()
	defer c.Unlock()

	return c.streamSessionID
}

func (c *apiClient) SetStreamSessionID(id string) {
	c.Lock()
	defer c.Unlock()

	c.streamSessionID = id
}

func (c *apiClient) writeJSON(v interface{}) error {
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	// err1 := json.NewEncoder(w).Encode(v)
	data, err1 := json.Marshal(v)
	logrus.WithField("endpoint", c.endpoint).Debugf("sending %s", string(data))
	_, err2 := w.Write(data)
	err3 := w.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return err3
}

func (c *apiClient) Connect(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()

	logrus.WithField("endpoint", c.endpoint).Debug("connecting")
	conn, res, err := websocket.DefaultDialer.DialContext(ctx, c.endpoint, nil)
	if err != nil {
		return err
	}

	logrus.WithField("endpoint", c.endpoint).Debug(*res)
	c.conn = conn

	if c.endpoint == "wss://ws.xapi.pro/demoStream" {
		c.state = Ready
	}
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

	logrus.WithField("endpoint", c.endpoint).Debug("logging in")
	msg := &loginMessage{
		Command: "login",
		Arguments: &LoginMessageArguments{
			UserID:   c.username,
			Password: c.password,
			AppName:  "kwont",
		},
	}

	c.state = AwaitingStreamSessionID

	if err := c.writeJSON(msg); err != nil {
		return err
	}

	return nil
}

func (c *apiClient) ReadMessages() error {
	defer func() {
		if err := c.conn.Close(); err != nil {
			logrus.WithField("endpoint", c.endpoint).WithError(err).Error("closing websocket connection")
		}
	}()

	logrus.WithField("endpoint", c.endpoint).Debug("reading messages")
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}

		message := string(data)
		logrus.WithField("endpoint", c.endpoint).Debug(message)

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
				logrus.WithField("endpoint", c.endpoint).Error(err)
				return err
			}
			c.streamSessionID = loginResponse.StreamSessionID
			c.state = Ready
		}
	}

	return nil
}

type SocketPingMessage struct {
	Command         string `json:"command"`
}

type StreamPingMessage struct {
	Command         string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
}

func (c *apiClient) SocketPingLoop() error {
	for range time.NewTicker(time.Second * 3).C {
		c.Lock()
		if c.state != Ready {
			c.Unlock()
			continue
		}

		if err := c.SocketPing(); err != nil {
			c.Unlock()
			return err
		}
		c.Unlock()
	}

	return nil
}

func (c *apiClient) SocketPing() error {
	logrus.WithField("endpoint", c.endpoint).Debug("socket pinging")

	return c.writeJSON(&SocketPingMessage{
		Command:         "ping",
	})
}

func (c *apiClient) StreamPingLoop() error {
	for range time.NewTicker(time.Second * 3).C {
		c.Lock()
		if c.state != Ready {
			c.Unlock()
			continue
		}

		if err := c.StreamPing(); err != nil {
			c.Unlock()
			return err
		}
		c.Unlock()
	}

	return nil
}

func (c *apiClient) StreamPing() error {
	logrus.WithField("endpoint", c.endpoint).Debug("stream pinging")

	return c.writeJSON(&StreamPingMessage{
		Command:         "ping",
		StreamSessionID: c.streamSessionID,
	})
}


type LoginResponse struct {
	Status          bool   `json:"status"`
	StreamSessionID string `json:"streamSessionId"`
}

type getTickPricesMessage struct {
	Command string `json:"command"`
	StreamSessionID string `json:"streamSessionId"`
	Symbol string `json:"symbol"`
	MinArrivalTime int `json:"minArrivalTime"`
	MaxLevel int `json:"maxLevel,omitempty"`
}


func (c *apiClient) StreamGetTickPrices(symbol string) error {
	c.Lock()
	defer c.Unlock()

	logrus.WithField("endpoint", c.endpoint).Debug("stream getting tick prices")

	if c.state != Ready {
		return errors.New("not ready")
	}

	return c.writeJSON(&getTickPricesMessage{
		Command: "getTickPrices",
		StreamSessionID: c.streamSessionID,
		Symbol: symbol,
		MinArrivalTime: 200,
		MaxLevel: 0,
	})
}


type GetNewsMessage struct {
Command string `json:"command"`
StreamSessionID string `json:"streamSessionId"`

}

func (c *apiClient) StreamGetNews() error {
	c.Lock()
	defer c.Unlock()

	logrus.WithField("endpoint", c.endpoint).Debug("stream getting news")

	if c.state != Ready {
		return errors.New("not ready")
	}

	return c.writeJSON(&GetNewsMessage{
		Command: "getNews",
		StreamSessionID: c.streamSessionID,
	})
}
