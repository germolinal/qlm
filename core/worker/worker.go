package main

import (
	"context"
	"log"
	"os"
	common "qml/core"
	"qml/core/ollamable"
	"strconv"
	"time"

	ollama "github.com/ollama/ollama/api"
	amqp "github.com/rabbitmq/amqp091-go"
)

var ollamaClient ollama.Client

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
	common.FailOnError(err, "Failed to declare a queue")

	genMessages, err := ch.Consume(
		genQueue.Name, // queue
		"",            // consumer
		false,         // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	common.FailOnError(err, "Failed to register a consumer")

	return genMessages
}

func listenMessages(ch *amqp.Channel) Queues {

	return Queues{
		chat:     listenQueue(ch, common.ChatQueue),
		generate: listenQueue(ch, common.GenerateQueue),
	}
}

func getConcurrency() int {
	name := "CONCURRENCY"
	concurrency := os.Getenv(name)
	if concurrency == "" {
		ret := 4
		log.Printf("no value for %s... using %d", name, ret)
		return ret
	}
	ret, err := strconv.Atoi(concurrency)
	if err != nil {
		log.Fatalf("value for '%s' is '%s' and cannot be turned into an int\n", name, concurrency)
	}
	return ret
}

func main() {

	// TODO: GATHER FROM OPTIONS

	concurrency := getConcurrency()
	//

	timeout := 600 * time.Second // LLMs are not always very fast... do not use a very small value here
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ollamaClient, err := ollama.ClientFromEnvironment()
	if err != nil {
		log.Fatal("Could not initialise ollama client")
	}

	connection := common.DialRabbit()
	defer connection.Close()

	ch, err := connection.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Limit concurrency
	err = ch.Qos(concurrency, 0, false)
	common.FailOnError(err, "Failed to set QoS")

	messages := listenMessages(ch)

	var forever chan struct{}

	go func() {
		for d := range messages.generate {
			ollamable.ProcessMsg(ctx, *&ollamaClient, common.GenerateQueue, d)
		}
		for d := range messages.chat {
			ollamable.ProcessMsg(ctx, *&ollamaClient, common.ChatQueue, d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
