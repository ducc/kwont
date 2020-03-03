package connections

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

type Connection struct {
	sync.RWMutex

	address string
	conn    *websocket.Conn
	closed  bool

	log *logrus.Entry
}

func New(ctx context.Context, address string) (*Connection, error) {
	log := logrus.WithField("address", address)
	log.Debug("connecting")

	conn, res, err := websocket.DefaultDialer.DialContext(ctx, address, nil)
	if err != nil {
		return nil, err
	}

	log.WithField("response", res).Debug("connected")

	return &Connection{
		address: address,
		conn:    conn,
		log:     log,
	}, nil
}

func (c *Connection) Closed() bool {
	c.RLock()
	defer c.RUnlock()

	return c.closed
}

var ErrConnectionClosed = errors.New("the connection has been closed")

func (c *Connection) WriteJSON(ctx context.Context, v interface{}) error {
	if c.Closed() {
		return ErrConnectionClosed
	}

	c.Lock()
	defer c.Unlock()

	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	data, err1 := json.Marshal(v)
	c.log.WithField("data", string(data)).Debugf("writing %s", string(data))

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

func (c *Connection) ReadMessages() <-chan []byte {
	messages := make(chan []byte)

	go func() {
		defer func() {
			if err := c.conn.Close(); err != nil {
				c.log.WithError(err).Error("closing connection")
			}

			close(messages)
		}()

		for {
			_, data, err := c.conn.ReadMessage()
			if err == io.EOF {
				c.log.Debug("message stream returned EOF")
				break
			}
			if err != nil {
				c.log.WithError(err).Error("reading messages")
				break
			}

			c.log.WithField("message", string(data)).Debug("received message")
			messages <- data
		}

		c.Lock()
		c.closed = true
		c.Unlock()
	}()

	return messages
}
