package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"orynt/internal/repository"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS bypass for local web sockets
	},
}

// Client represents a connected user/session websocket
type Client struct {
	hub  *WSHub
	conn *websocket.Conn
	send chan []byte
}

// WSHub maintains active clients and broadcasts updates
type WSHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

// NewWSHub creates a WS manager
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run listens for registry and broadcast events
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client connected. Total active: %d", len(h.clients))
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WS] Client disconnected. Total active: %d", len(h.clients))
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// In Orynt stadium, clients mostly listen to telemetry (unidirectional flows).
		// Interactive updates (like task status change, incident report, chat) are sent via standard REST POST/PUT requests.
		// So we ignore incoming messages from client here.
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// WebSocketHandler manages upgrading connections
func WebSocketHandler(hub *WSHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WS] Upgrade error: %v", err)
			return
		}
		client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}

// StartPubSubListener links Redis/Memory PubSub to WebSockets Hub
func StartPubSubListener(pubSub repository.PubSubRepository, hub *WSHub) {
	channels := []string{"stadium", "parking", "transport", "alerts", "announcements", "tasks", "lost_found", "medical", "matches"}

	for _, ch := range channels {
		c := ch
		_ = pubSub.Subscribe(context.Background(), c, func(msg string) {
			// Forward string message payload directly to websocket broadcast channel
			hub.broadcast <- []byte(msg)
		})
	}
}

// PublishWSUtility parses and publishes to a channel
func PublishWSUtility(pubSub repository.PubSubRepository, channel string, messageType string, payload interface{}) {
	msg := struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}{
		Type:    messageType,
		Payload: payload,
	}
	bytes, err := json.Marshal(msg)
	if err == nil {
		_ = pubSub.Publish(context.Background(), channel, string(bytes))
	}
}
