package main

import (
	"fmt"
	"log"

	"github.com/jarimus/learn-pub-sub-starter/internal/gamelogic"
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
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("error opening channel: %v", err)
	}
	fmt.Println("Channel opened.")

	// Subscribe to game logs
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		handlerGameLogs(),
	)
	if err != nil {
		log.Fatalf("error subscriging to game logs: %v", err)
	}

	// Server Help
	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		firstWord := words[0]
		switch firstWord {
		case "pause":
			fmt.Println("Sendind a pause message.")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("error publishing: %v", err)
			}
		case "resume":
			fmt.Println("Sending a resume message.")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("error publishing: %v", err)
			}
		case "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Print("Invalid command.\nI understand 'pause', 'resume' and 'quit'.")
		}
	}
}
