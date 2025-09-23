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
	testChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("error opening channel: %v", err)
	}
	fmt.Println("Channel opened.")

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
				testChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
		case "resume":
			fmt.Println("Sending a resume message.")
			pubsub.PublishJSON(
				testChannel,
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

	// // Waiting for ctrl+c
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	fmt.Println("\nShutting down...\nClosing connection...")
}
