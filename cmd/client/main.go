package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	conAddr := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(conAddr)
	if err != nil {
		fmt.Println("connection rabbitmq issue: " + err.Error())
		return
	}

	defer connection.Close()
	fmt.Println("Connection successful")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Can not welcome client: %w", err)
		return
	}

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient)

	chanl, err := connection.Channel()
	if err != nil {
		fmt.Println("can not create a chanel")
		return
	}

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(connection,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState))

	if err != nil {
		fmt.Println(err)
		return
	}

	err = pubsub.SubscribeJSON(connection,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gameState))

	if err != nil {
		fmt.Println(err)
		return
	}

Loop:
	for {
		command := gamelogic.GetInput()
		if len(command) == 0 {
			continue
		}
		switch command[0] {
		case "spawn":
			if err := gameState.CommandSpawn(command); err != nil {
				fmt.Println(err)
			}
		case "move":
			move, err := gameState.CommandMove(command)
			if err != nil {
				fmt.Println(err)
			}

			pubsub.PublishJSON(chanl, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+username, move)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break Loop
		default:
			fmt.Println("Unknown command")
		}
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(mv gamelogic.ArmyMove) {
		gs.HandleMove(mv)
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		gs.HandlePause(ps)
	}
}
