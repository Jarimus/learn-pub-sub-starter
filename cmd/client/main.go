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
	fmt.Println("Starting Peril client...")
	// Connect
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Client connected to RabbitMQ!")

	// Open channel
	publishChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("error opening channel: %v", err)
	}
	defer publishChannel.Close()

	// Welcome message
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error setting username: %v", err)
	}

	// New game state
	gamestate := gamelogic.NewGameState(username)

	// Subscribe to pauses
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gamestate.GetUsername(),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gamestate),
	)
	if err != nil {
		fmt.Printf("error subscribing to pause: %v", err)
	}

	// Subscribe to player moves
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gamestate, publishChannel),
	)
	if err != nil {
		fmt.Printf("error subscribing to player moves: %v", err)
	}

	// Subscribe to wars
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(gamestate, publishChannel),
	)
	if err != nil {
		fmt.Printf("error subscribing to wars: %v", err)
	}

	// REPL
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		firstWord := words[0]
		switch firstWord {
		case "spawn":
			err = gamestate.CommandSpawn(words)
			if err != nil {
				log.Printf("%v", err)
			}

		case "move":
			armyMove, err := gamestate.CommandMove(words)
			if err != nil {
				log.Printf("%v", err)
				break
			}
			err = pubsub.PublishJSON(
				publishChannel,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				armyMove,
			)
			if err != nil {
				fmt.Printf("error publishing move: %v", err)
				break
			}
			fmt.Println("move published successfully")
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Invalid command.")
		}
	}
}
