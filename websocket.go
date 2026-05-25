package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 // We only send small control messages or JSON from client
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for MVP
	},
}

// Client represents an active WebSocket connection
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID int
	send   chan []byte
}

// Hub maintains the set of active clients
type Hub struct {
	// Registered clients mapped by userID to allow sending to specific users
	clients map[int]map[*Client]bool
	
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var globalHub *Hub

func newHub() *Hub {
	return &Hub{
		clients:    make(map[int]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID][client]; ok {
				delete(h.clients[client.userID], client)
				close(client.send)
				if len(h.clients[client.userID]) == 0 {
					delete(h.clients, client.userID)
				}
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			// For global broadcasts if needed
			h.mu.RLock()
			for _, userClients := range h.clients {
				for client := range userClients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(userClients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser sends a message to all active WebSocket connections for a specific user
func (h *Hub) SendToUser(userID int, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if userClients, ok := h.clients[userID]; ok {
		for client := range userClients {
			select {
			case client.send <- message:
			default:
				// If send buffer is full, remove the client
				close(client.send)
				delete(userClients, client)
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// MVP: We only receive messages via REST API, WS is for server->client pushes.
		// If you want clients to send messages via WS, process them here.
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate the user before upgrading
	userID, ok := getUserIDFromSession(r)
	if !ok {
		// Try to get token from URL if cookies aren't sent with WS connection
		token := r.URL.Query().Get("token")
		if token != "" {
			var dbUserID int
			err := DB.QueryRow(`SELECT user_id FROM sessions WHERE session_token = ? AND expires_at > ?`, token, time.Now()).Scan(&dbUserID)
			if err == nil {
				userID = dbUserID
				ok = true
			}
		}
	}
	
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	
	client := &Client{
		hub:    globalHub,
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// WSMessage represents a message sent to the client
type WSMessage struct {
	Type    string      `json:"type"`    // e.g., "new_message", "status_update"
	Payload interface{} `json:"payload"` // the actual data
}

// NotifyUser sends a typed JSON message to a user via WebSocket
func NotifyUser(userID int, msgType string, payload interface{}) {
	if globalHub == nil {
		return
	}
	
	msg := WSMessage{
		Type:    msgType,
		Payload: payload,
	}
	
	data, err := json.Marshal(msg)
	if err == nil {
		globalHub.SendToUser(userID, data)
	}
}
