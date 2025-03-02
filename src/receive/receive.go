package main

import (
	"context"
	"encoding/json"
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

func genGenCallback(req *ollama.GenerateRequest) ollama.GenerateResponseFunc {
	return func(res ollama.GenerateResponse) error {
		log.Printf("%s\n", res.Response)
		return nil
	}
}

func main() {

	/*
		Options to be configured:
		- model
	*/
	generateTimeout := 30 * time.Second

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

	genQueue, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
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

	var forever chan struct{}

	go func() {

		// Process Generate
		for d := range genMessages {
			var req ollama.GenerateRequest
			err := json.Unmarshal(d.Body, &req)
			if err != nil {
				// todo: This was a bad request... will not be fixed by retries.
				log.Fatal("bad request in Generate: ", string(d.Body))
			}

			// Default values
			stream := false
			req.Stream = &stream
			ctx, _ := context.WithTimeout(context.Background(), generateTimeout)
			err = ollamaClient.Generate(ctx, &req, genGenCallback(&req))
			if err != nil {
				// TODO: Generation failed... this might be fixed by retrying
				log.Fatal("ollama failed: ", string(d.Body))
				d.Nack(false, true)
			}
			log.Printf("Received a message: %s", d.Body)
			d.Ack(false)

		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
