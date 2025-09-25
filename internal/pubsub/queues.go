package pubsub

import (
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Transient SimpleQueueType = "transient"
	Durable   SimpleQueueType = "durable"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
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

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueType,
	)
	if err != nil {
		return err
	}
	if queue.Name == "" {
		return fmt.Errorf("cannot subscribe to queue: queue does not exist")
	}
	msgs, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error opening delivery channel: %v", err)
	}
	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}

	go func() {
		defer channel.Close()
		for msg := range msgs {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			switch handler(target) {
			case Ack:
				// fmt.Print("Acked.")
				msg.Ack(false)
			case NackRequeue:
				// fmt.Print("NackRequeued.")
				msg.Nack(false, true)
			case NackDiscard:
				// fmt.Print("NackDiscarded.")
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}
