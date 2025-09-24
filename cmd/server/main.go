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
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("error opening channel: %v", err)
	}
	fmt.Println("Channel opened.")

	// Declare and bind
	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.Durable,
	)
	if err != nil {
		log.Fatalf("error declaring and binding queue to exchange: %v", err)
	}
	fmt.Printf("queue declared and bound: %v.\n", queue.Name)

	// Server Help
	gamelogic.PrintServerHelp()

	looping := true
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		firstWord := words[0]
		switch firstWord {
		case "pause":
			fmt.Println("Sendind a pause message.")
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
		case "resume":
			fmt.Println("Sending a resume message.")
			pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
		case "quit":
			fmt.Println("Exiting...")
			looping = false
		default:
			fmt.Print("Invalid command.\nI understand 'pause', 'resume' and 'quit'.")
		}
		if !looping {
			break
		}
	}

	fmt.Println("\nShutting down...\nClosing connection...")
}
