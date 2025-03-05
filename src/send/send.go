package main

import (
	"context"
	"encoding/json"
	common "gocheck"
	"gocheck/ollamable"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Queues struct {
	chat     amqp.Queue
	generate amqp.Queue
}

// func queue(ch *amqp.Channel, name string) amqp.Queue {
// 	q, err := ch.QueueDeclare(
// 		name,  // name
// 		false, // durable
// 		false, // delete when unused
// 		false, // exclusive
// 		false, // no-wait
// 		nil,   // arguments
// 	)
// 	failOnError(err, "Failed to declare a queue")
// 	return q
// }

// func queues(ch *amqp.Channel) Queues {
// 	return Queues{
// 		chat:     queue(ch, common.ChatQueue),
// 		generate: queue(ch, common.GenerateQueue),
// 	}
// }

func publishRequest(ctx context.Context, ch *amqp.Channel, queueName string, body []byte) {
	err := ch.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to Rabbit MQ
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connection.Close()

	ch, err := connection.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/api/chat", func(c *gin.Context) {
		var req ollamable.LLMChatRequest
		// Check that everything is working fine
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		body, err := json.Marshal(req)
		failOnError(err, "error marshalling chat request?")
		publishRequest(ctx, ch, common.ChatQueue, body)

	})
	r.POST("/api/generate", func(c *gin.Context) {
		var req ollamable.LLMGenerateRequest
		// Check that everything is working fine
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		body, err := json.Marshal(req)
		failOnError(err, "error marshalling chat request?")
		publishRequest(ctx, ch, common.GenerateQueue, body)

	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
