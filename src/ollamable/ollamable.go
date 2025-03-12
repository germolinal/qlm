package ollamable

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	common "gocheck"
	"log"
	"net/http"
	"time"

	ollama "github.com/ollama/ollama/api"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Ollamable interface {
	SetModel(string)
	SetStream(*bool)
	GetHook() string
}

type LLMChatRequest struct {
	ollama.ChatRequest
	Hook string `json:"webhook"`
}

func (c *LLMChatRequest) SetModel(m string) {
	c.Model = m
}

func (c *LLMChatRequest) SetStream(m *bool) {
	c.Stream = m
}
func (c *LLMChatRequest) GetHook() string {
	return c.Hook
}

type LLMGenerateRequest struct {
	ollama.GenerateRequest
	Hook string `json:"webhook"`
}

func (c *LLMGenerateRequest) SetModel(m string) {
	c.Model = m
}

func (c *LLMGenerateRequest) SetStream(m *bool) {
	c.Stream = m
}
func (c *LLMGenerateRequest) GetHook() string {
	return c.Hook
}

func hook(url string, msg []byte) error {
	log.Printf("Attempting to POST message to URL: %s\nMessage: %s\n", url, string(msg))

	// 1. Create an HTTP client (you can reuse a client for better performance if calling hook frequently)
	client := &http.Client{
		Timeout: time.Second * 10, // Set a timeout for the request (e.g., 10 seconds)
	}

	// 2. Construct the POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(msg))
	if err != nil {
		log.Printf("Error creating POST request: %v", err)
		return err // Return the error to the caller
	}

	// Set Content-Type header to application/json (assuming your webhook expects JSON)
	req.Header.Set("Content-Type", "application/json")

	// 3. Execute the POST request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing POST request: %v", err)
		return err // Return the error to the caller
	}
	defer resp.Body.Close() // Ensure response body is closed after function returns

	// 4. Handle response and check for errors
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Successfully POSTed message to URL: %s, Status Code: %d\n", url, resp.StatusCode)
		return nil // Success, return nil error
	} else {
		log.Printf("POST request to URL: %s failed with Status Code: %d\n", url, resp.StatusCode)
		// Optionally, you could read the response body here to log error details from the webhook receiver
		// bodyBytes, _ := io.ReadAll(resp.Body)
		// log.Printf("Response Body: %s\n", string(bodyBytes))
		return fmt.Errorf("POST request failed with status code: %d", resp.StatusCode) // Return error for non-successful status codes
	}
}

func ProcessMsg(ctx context.Context, ollamaClient *ollama.Client, queueName string, d amqp.Delivery) {
	fmt.Println("received!")
	var req interface{} // Use interface{} to handle different request types

	switch queueName {
	case common.GenerateQueue:
		req = &LLMGenerateRequest{}
	case common.ChatQueue:
		req = &LLMChatRequest{}
	default:
		log.Println("unsupported queue name")
		d.Nack(false, false)
		return
	}

	err := json.Unmarshal(d.Body, &req)
	if err != nil {
		log.Printf("bad request: %v", err)
		d.Nack(false, false)
		return
	}

	// Default values
	stream := false
	if ollamaReq, ok := req.(Ollamable); ok {
		ollamaReq.SetStream(&stream)
		ollamaReq.SetModel("gemma2")
	} else {
		log.Println("request is not OllamaRequest")
		d.Nack(false, false)
		return
	}

	hookUrl := req.(Ollamable).GetHook()
	switch r := req.(type) {
	case *LLMGenerateRequest:
		err = ollamaClient.Generate(ctx, &r.GenerateRequest, func(res ollama.GenerateResponse) error {
			s, err := json.Marshal(res)
			if err != nil {
				return err
			}
			return hook(hookUrl, s)
		})
		if err != nil {
			log.Printf("ollama generate failed: %v", err)
			d.Nack(false, true)
			return
		}
	case *LLMChatRequest:
		err = ollamaClient.Chat(ctx, &r.ChatRequest, func(res ollama.ChatResponse) error {
			s, err := json.Marshal(res)
			if err != nil {
				return err
			}
			return hook(hookUrl, s)
		})
		if err != nil {
			log.Printf("ollama chat failed: %v", err)
			d.Nack(false, true)
			return
		}
	default:
		log.Println("unknown request type after unmarshal")
		d.Nack(false, false)
		return
	}

	log.Printf("Received a message: %s", d.Body)
	d.Ack(false)
}
