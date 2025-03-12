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
		return true // Allow connections from any origin (for development - IMPORTANT: adjust for production)
	},
}

// Client represents a connected websocket client.
type Client struct {
	conn *websocket.Conn
}

// Global map to store connected clients (using a map for easy removal and concurrency safety)
var clients = make(map[*Client]bool)
var clientsMutex sync.Mutex // Mutex to protect concurrent access to clients map

// broadcastMessage sends a message to all connected clients.
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
		_, _, err := conn.ReadMessage() // Keep connection alive, but we are server-push only in this example
		if err != nil {
			clientsMutex.Lock()
			delete(clients, client)
			clientsMutex.Unlock()
			log.Println("Client disconnected")
			break
		}
		// We are not expecting messages from the client in this example, but you could handle them here if needed.
		// For example, you could echo back or process client messages.
		// messageType, p, err := conn.ReadMessage()
		// if err != nil {
		// 	log.Printf("error reading message: %v", err)
		// 	break
		// }
		// log.Printf("Received message: %s", p)
		// if err := conn.WriteMessage(messageType, p); err != nil {
		// 	log.Printf("error writing message: %v", err)
		// 	break
		// }
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
