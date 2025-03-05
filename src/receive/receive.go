package main

import (
	"context"
	"fmt"
	common "gocheck"
	"gocheck/ollamable"
	"log"
	"time"

	ollama "github.com/ollama/ollama/api"
	amqp "github.com/rabbitmq/amqp091-go"
)

var ollamaClient ollama.Client

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Queues struct {
	chat     <-chan amqp.Delivery
	generate <-chan amqp.Delivery
}

func listenQueue(ch *amqp.Channel, name string) <-chan amqp.Delivery {
	genQueue, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	genMessages, err := ch.Consume(
		genQueue.Name, // queue
		"",            // consumer
		false,         // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	failOnError(err, "Failed to register a consumer")

	return genMessages
}

func listenMessages(ch *amqp.Channel) Queues {

	return Queues{
		chat:     listenQueue(ch, common.ChatQueue),
		generate: listenQueue(ch, common.GenerateQueue),
	}
}

func main() {

	/*
		Options to be configured:
		- model
	*/
	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	/* END OF PARSE OPTIONS */
	ollamaClient, err := ollama.ClientFromEnvironment()
	if err != nil {
		log.Fatal("Could not initialise ollama client")
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	messages := listenMessages(ch)

	var forever chan struct{}

	go func() {
		for d := range messages.generate {
			fmt.Println("gen message")
			ollamable.ProcessMsg(ctx, *&ollamaClient, common.GenerateQueue, d)
		}
		for d := range messages.chat {
			fmt.Println("chat message")
			ollamable.ProcessMsg(ctx, *&ollamaClient, common.ChatQueue, d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
