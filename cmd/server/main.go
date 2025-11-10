package main

import (
	"fmt"
	"os"
	"os/signal"

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

	chanl, err := connection.Channel()
	if err != nil {
		fmt.Println("can not create a chanel")
		return
	}

	pubsub.PublishJSON(chanl, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	fmt.Println("close the connection")
	connection.Close()
}
