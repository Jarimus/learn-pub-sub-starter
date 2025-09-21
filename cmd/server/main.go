package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jarimus/learn-pub-sub-starter/internal/pubsub"
	"github.com/jarimus/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Connect
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected successfully...")
	// New channel
	testChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("error opening channel: %v", err)
	}

	pubsub.PublishJSON(
		testChannel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		},
	)

	// Waiting for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nShutting down...\nClosing connection...")
}
