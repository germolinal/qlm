package common

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const GenerateQueue string = "generate"
const ChatQueue string = "chat"

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func DialRabbit() *amqp.Connection {
	connection, err := amqp.Dial("amqp://guest:guest@rabbit:5672/")
	FailOnError(err, "Failed to connect to RabbitMQ")
	return connection
}
