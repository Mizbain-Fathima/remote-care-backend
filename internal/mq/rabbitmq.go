package mq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	exch string
}

func NewPublisher(url, exchange string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, err
	}
	return &Publisher{conn: conn, ch: ch, exch: exchange}, nil
}

func (p *Publisher) PublishPurchase(ctx context.Context, routingKey string, body []byte) error {
	return p.ch.PublishWithContext(ctx,
		p.exch,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Publisher) Close() {
	p.ch.Close()
	p.conn.Close()
}
