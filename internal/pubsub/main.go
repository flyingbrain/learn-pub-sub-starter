package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("Can not marshal to json %e", err)
	}

	ctx := context.Background()
	publish := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}

	if err := ch.PublishWithContext(ctx, exchange, key, false, false, publish); err != nil {
		return fmt.Errorf("can not publish contest to chennel %e", err)
	}

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	chanl, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("can not create a chanel %w", err)
	}

	q, err := chanl.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType == Transient,
		queueType == Transient,
		false,
		nil)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("can't declare queue %w", err)
	}

	err = chanl.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("can't bind queue %w", err)
	}

	return chanl, q, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T),
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("can't declare the queue %w", err)
	}

	chanl, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("can't create Channel %w", err)
	}
	go func() {
		for message := range chanl {
			var mess T

			if err := json.Unmarshal(message.Body, &mess); err != nil {
			}
			handler(mess)
			message.Ack(false)
			fmt.Print("> ")
		}
	}()

	return nil
}
