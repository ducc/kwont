package pubsub

import (
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
)

var amqpAddress string

func init() {
	flag.StringVar(&amqpAddress, "amqp-address", "", "amqp server connection address")
}

type Client struct {
	conn   *amqp.Connection
	mutex  *sync.Mutex
	queues []*Queue
}

func New() (*Client, error) {
	conn, err := amqp.Dial(amqpAddress)
	if err != nil {
		logrus.WithError(err).Fatal("connecting to amqp server")
	}

	return &Client{
		conn:   conn,
		mutex:  &sync.Mutex{},
		queues: make([]*Queue, 0),
	}, nil
}

func (c *Client) Close() error {
	for _, queue := range c.queues {
		queue.close()
	}
	return c.conn.Close()
}

type Queue struct {
	topic   string
	channel *amqp.Channel
	queue   amqp.Queue
}

func (q *Queue) Subscribe() (<-chan *Message, error) {
	deliveries, err := q.channel.Consume(
		q.queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	messages := make(chan *Message)
	go func() {
		defer close(messages)
		for delivery := range deliveries {
			messages <- &Message{
				delivery: delivery,
			}
		}
	}()

	return messages, nil
}

func (q *Queue) Publish(body []byte) error {
	return q.channel.Publish(
		"",
		q.queue.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (q *Queue) close() {
	if err := q.channel.Close(); err != nil {
		logrus.WithError(err).WithField("topic", q.topic).Error("closing queue")
	}
}

func (c *Client) Queue(topic string) (*Queue, error) {
	channel, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		topic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	q := &Queue{topic: topic, channel: channel, queue: queue}
	c.queues = append(c.queues, q)

	return q, nil
}

type Message struct {
	delivery amqp.Delivery
}

func (m *Message) Body() []byte {
	return m.delivery.Body
}

func (m *Message) Ack() error {
	return m.delivery.Ack(false)
}
