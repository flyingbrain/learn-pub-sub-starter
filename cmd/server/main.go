package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	conAddr := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(conAddr)
	if err != nil {
		fmt.Println("connection rabbitmq issue: " + err.Error())
		return
	}

	defer connection.Close()

	fmt.Println("Connection successful")

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable)

	chanl, err := connection.Channel()
	if err != nil {
		fmt.Println("can not create a chanel")
		return
	}

	gamelogic.PrintServerHelp()

Loop:
	for {
		command := gamelogic.GetInput()
		if len(command) == 0 {
			continue
		}

		switch command[0] {
		case "pause":
			pubsub.PublishJSON(chanl, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			fmt.Println("game is on pause")
		case "resume":
			pubsub.PublishJSON(chanl, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			fmt.Println("game unpaused")
		case "quit":
			fmt.Println("quit the game")
			break Loop
		default:
			fmt.Println("Unknown command")
		}
	}

	pubsub.PublishJSON(chanl, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	fmt.Println("close the connection")
	connection.Close()
}
