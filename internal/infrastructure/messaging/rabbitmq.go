package messaging

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer consumes messages from RabbitMQ queues bound to topic exchanges.
type Consumer struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	prefetch int
	mu       sync.Mutex
}

// NewConsumer creates a new consumer.
func NewConsumer(url string, prefetch int) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	return &Consumer{conn: conn, ch: ch, prefetch: prefetch}, nil
}

// EnsureQueue declares the topic exchange, queue, and bind.
func (c *Consumer) EnsureQueue(topicName, queueName, routingKey string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ch.ExchangeDeclare(topicName, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	_, err := c.ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}
	if routingKey == "" {
		routingKey = "#"
	}
	return c.ch.QueueBind(queueName, routingKey, topicName, false, nil)
}

// Consume runs the handler for each delivery until ctx is cancelled.
func (c *Consumer) Consume(ctx context.Context, queueName string, handler func(ctx context.Context, body []byte, delivery amqp.Delivery)) error {
	deliveries, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}
			handler(ctx, d.Body, d)
		}
	}
}

// Close closes channel and connection.
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

var _ = time.Sleep
