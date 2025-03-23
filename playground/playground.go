package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin
		// (This is a playground... adjust for production)
		return true
	},
}

// Client represents a connected websocket client.
type Client struct {
	conn *websocket.Conn
}

var client Client

func sendMsgToUI(message interface{}) {

	err := client.conn.WriteJSON(message)
	if err != nil {
		log.Printf("error sending message to client: %v", err)
		client.conn.Close()
	}

}

// handleConnections upgrades HTTP connection to WebSocket and manages clients.
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	client = Client{conn: conn}

	log.Println("Client connected")

	for {
		// Keep connection alive
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Client disconnected")
			break
		}

	}
}

// handleWebhook receives POST requests and broadcasts the message to WebSocket clients.
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var messagePayload interface{} // Use interface{} to handle various JSON structures
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&messagePayload)
	if err != nil {
		http.Error(w, "Error decoding JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	sendMsgToUI(messagePayload)
}

// handles a message sent from the front end
func handleMsg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var messagePayload interface{} // Use interface{} to handle various JSON structures
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&messagePayload)
	if err != nil {
		http.Error(w, "Error decoding JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Turn into bytes
	payloadBytes, err := json.Marshal(messagePayload)
	if err != nil {
		http.Error(w, "Error encoding JSON payload", http.StatusInternalServerError)
		return
	}

	// Create request
	orchestratoUrl := "http://localhost:8080/api/generate"
	fowardReq, err := http.NewRequest(http.MethodPost, orchestratoUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, "Error creating request to other server", http.StatusInternalServerError)
		return
	}
	fowardReq.Header.Set("Content-Type", "application/json")

	// Forward request to orchestrator
	client := &http.Client{}
	otherResp, err := client.Do(fowardReq)
	if err != nil {
		fmt.Println(".... wrong here")
		http.Error(w, "Error sending request to other server", http.StatusInternalServerError)
		return
	}
	defer otherResp.Body.Close()
	// Return whatever the orchestrator responsed
	w.WriteHeader(otherResp.StatusCode)
	_, err = io.Copy(w, otherResp.Body)
	if err != nil {
		fmt.Println("wrong 2")
		log.Fatal(err)
		return
	}
	fmt.Fprint(w, "Webhook message received and broadcasted")
}

func main() {

	fs := http.FileServer(http.Dir("./ui"))
	http.Handle("/", fs)

	http.HandleFunc("/socket", handleConnections)
	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/msg", handleMsg)

	fmt.Println("WebSocket server starting on :3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
