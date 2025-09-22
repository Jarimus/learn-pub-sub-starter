package pubsub

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType string

const (
	Transient simpleQueueType = "transient"
	Durable   simpleQueueType = "durable"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	// open channel
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// Declare queue
	var queueTypeBool bool
	switch queueType {
	case Transient:
		queueTypeBool = false
	case Durable:
		queueTypeBool = true
	default:
		return nil, amqp.Queue{}, errors.New("invalid queue type")
	}
	queue, err := channel.QueueDeclare(queueName, queueTypeBool, !queueTypeBool, !queueTypeBool, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// Bind queue to exchange
	err = channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}
