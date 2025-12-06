package mq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewConsumer(url, exchange, routingKey string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// must declare same exchange as publisher
	err = ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// create queue
	q, err := ch.QueueDeclare(
		"purchase_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// bind queue to exchange
	err = ch.QueueBind(
		q.Name,
		routingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{conn, ch, q.Name}, nil
}

// Consume messages with a handler function
func (c *Consumer) Consume(ctx context.Context, handler func([]byte) error) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		true, // auto-ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume failed: %w", err)
	}

	for {
		select {
		case msg := <-msgs:
			if len(msg.Body) == 0 {
				continue
			}
			if err := handler(msg.Body); err != nil {
				fmt.Println("handler error:", err)
			}

		case <-ctx.Done():
			return c.conn.Close()
		}
	}
}
