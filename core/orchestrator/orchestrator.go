package main

import (
	"context"
	"encoding/json"
	"net/http"
	common "qml/core"
	"qml/core/ollamable"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Queues struct {
	chat     amqp.Queue
	generate amqp.Queue
}

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
	common.FailOnError(err, "Failed to publish a message")
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to Rabbit MQ
	connection := common.DialRabbit()
	defer connection.Close()

	ch, err := connection.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	r := gin.Default()

	r.POST("/api/chat", func(c *gin.Context) {
		var req ollamable.LLMChatRequest
		// Check that everything is working fine
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		body, err := json.Marshal(req)
		common.FailOnError(err, "error marshalling chat request?")
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
		common.FailOnError(err, "error marshalling chat request?")
		publishRequest(ctx, ch, common.GenerateQueue, body)

	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
