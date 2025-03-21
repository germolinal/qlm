package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

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

var clients = make(map[*Client]bool)

// Mutex to protect concurrent access to clients map
var clientsMutex sync.Mutex

func broadcastMessage(message interface{}) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		err := client.conn.WriteJSON(message)
		if err != nil {
			log.Printf("error broadcasting message to client: %v", err)
			client.conn.Close()
			delete(clients, client)
		}
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

	client := &Client{conn: conn}
	clientsMutex.Lock()
	clients[client] = true
	clientsMutex.Unlock()

	log.Println("Client connected")

	for {
		// Keep connection alive
		_, _, err := conn.ReadMessage()
		if err != nil {
			clientsMutex.Lock()
			delete(clients, client)
			clientsMutex.Unlock()
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

	log.Printf("Received webhook message: %+v", messagePayload)
	broadcastMessage(messagePayload)

	fmt.Fprint(w, "Webhook message received and broadcasted") // Respond to the webhook sender
}

func main() {

	clients = make(map[*Client]bool) // Initialize the clients map

	http.HandleFunc("/socket", handleConnections)
	http.HandleFunc("/webhook", handleWebhook)

	fmt.Println("WebSocket server starting on :3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
