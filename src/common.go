package common

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

const GenerateQueue string = "generate"
const ChatQueue string = "chat"

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func EnvOrFail(name string) string {
	ret := os.Getenv(name)
	if ret == "" {
		log.Fatalf("could not find '%s'", name)
	}
	return ret
}

func rabbitUrl() string {
	url := EnvOrFail("RABBIT_URL")
	username := EnvOrFail("RABBIT_USERNAME")
	password := EnvOrFail("RABBIT_PASSWORD")
	return fmt.Sprintf("amqp://%s:%s@%s/", username, password, url)
}

func DialRabbit() *amqp.Connection {
	url := rabbitUrl()
	connection, err := amqp.Dial(url)
	FailOnError(err, "Failed to connect to RabbitMQ")
	return connection
}
