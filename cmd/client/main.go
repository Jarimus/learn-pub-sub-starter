package main

import (
	"fmt"
	"log"
	"strconv"

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

	// Welcome message
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error setting username: %v", err)
	}

	// Declare and bing the pause key
	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("error declaring and binding queue to exchange: %v", err)
	}
	fmt.Printf("queue declared and bound: %v", queue.Name)

	// New game state
	gamestate := gamelogic.NewGameState(username)

	// REPL
	looping := true
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		firstWord := words[0]
		switch firstWord {
		case "spawn":
			err := gamestate.CommandSpawn(words)
			if err != nil {
				log.Printf("%v", err)
			}

		case "move":
			armyMove, err := gamestate.CommandMove(words)
			if err != nil {
				log.Printf("%v", err)
			}
			var units string
			for i, unit := range armyMove.Units {
				if i < len(armyMove.Units)-1 {
					units += strconv.Itoa(unit.ID) + ", "
				} else {
					if len(armyMove.Units) > 1 {
						units += "and " + strconv.Itoa(unit.ID)
					} else {
						units += strconv.Itoa(unit.ID)
					}
				}
			}
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed!")
		case "quit":
			gamelogic.PrintQuit()
			looping = false
		default:
			fmt.Println("Invalid command.")
		}
		if !looping {
			break
		}
	}

	// wait for ctrl+c
	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, os.Interrupt)
	// <-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
