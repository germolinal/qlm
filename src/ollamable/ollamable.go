package ollamable

import (
	"context"
	"encoding/json"
	"fmt"
	common "gocheck"
	"log"

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
	Hook string `json:"hook"`
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
	Hook string `json:"hook"`
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
	log.Printf("%s\n", string(msg))
	return nil
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
