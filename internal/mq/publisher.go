package mq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
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

	// Declare topic exchange
	err = ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
	}, nil
}

func (p *Publisher) PublishPurchase(ctx context.Context, routingKey string, body []byte) error {
	return p.channel.PublishWithContext(
		ctx,
		p.exchange,
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
	p.channel.Close()
	p.conn.Close()
}
