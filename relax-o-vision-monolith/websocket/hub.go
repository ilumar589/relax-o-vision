package websocket

import (
	"log/slog"
	"sync"

	fiberws "github.com/gofiber/contrib/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	Conn          *fiberws.Conn
	Send          chan *WSMessage
	Subscriptions map[string]bool // rooms the client is subscribed to
	mu            sync.RWMutex
}

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Client subscriptions organized by room
	rooms map[string]map[*Client]bool

	// Broadcast channel for messages to all clients
	broadcast chan *WSMessage

	// Room-specific broadcast
	roomBroadcast chan *RoomMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Subscribe requests
	subscribe chan *SubscribeRequest

	// Unsubscribe requests
	unsubscribe chan *UnsubscribeRequest

	mu sync.RWMutex
}

// RoomMessage represents a message to be sent to a specific room
type RoomMessage struct {
	Room    string
	Message *WSMessage
}

// SubscribeRequest represents a client subscription request
type SubscribeRequest struct {
	Client *Client
	Room   string
}

// UnsubscribeRequest represents a client unsubscription request
type UnsubscribeRequest struct {
	Client *Client
	Room   string
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		broadcast:     make(chan *WSMessage, 256),
		roomBroadcast: make(chan *RoomMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscribe:     make(chan *SubscribeRequest),
		unsubscribe:   make(chan *UnsubscribeRequest),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case roomMsg := <-h.roomBroadcast:
			h.broadcastToRoom(roomMsg.Room, roomMsg.Message)

		case sub := <-h.subscribe:
			h.subscribeToRoom(sub.Client, sub.Room)

		case unsub := <-h.unsubscribe:
			h.unsubscribeFromRoom(unsub.Client, unsub.Room)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	slog.Info("Client registered", "clientId", client.ID)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)

		// Remove from all rooms
		for room := range client.Subscriptions {
			if clients, ok := h.rooms[room]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.rooms, room)
				}
			}
		}

		slog.Info("Client unregistered", "clientId", client.ID)
	}
}

// broadcastToAll sends a message to all connected clients
func (h *Hub) broadcastToAll(message *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			slog.Warn("Failed to send message to client", "clientId", client.ID)
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

// broadcastToRoom sends a message to all clients in a specific room
func (h *Hub) broadcastToRoom(room string, message *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		for client := range clients {
			select {
			case client.Send <- message:
			default:
				slog.Warn("Failed to send message to client in room", "clientId", client.ID, "room", room)
			}
		}
	}
}

// subscribeToRoom subscribes a client to a room
func (h *Hub) subscribeToRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	client.Subscriptions[room] = true
	client.mu.Unlock()

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true

	slog.Info("Client subscribed to room", "clientId", client.ID, "room", room)

	// Send confirmation to client
	msg, _ := NewMessage(EventSubscribed, map[string]string{"room": room})
	select {
	case client.Send <- msg:
	default:
	}
}

// unsubscribeFromRoom unsubscribes a client from a room
func (h *Hub) unsubscribeFromRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.mu.Lock()
	delete(client.Subscriptions, room)
	client.mu.Unlock()

	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}

	slog.Info("Client unsubscribed from room", "clientId", client.ID, "room", room)

	// Send confirmation to client
	msg, _ := NewMessage(EventUnsubscribed, map[string]string{"room": room})
	select {
	case client.Send <- msg:
	default:
	}
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(message *WSMessage) {
	h.broadcast <- message
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(room string, message *WSMessage) {
	h.roomBroadcast <- &RoomMessage{
		Room:    room,
		Message: message,
	}
}

// Register registers a new client
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Subscribe subscribes a client to a room
func (h *Hub) Subscribe(client *Client, room string) {
	h.subscribe <- &SubscribeRequest{
		Client: client,
		Room:   room,
	}
}

// Unsubscribe unsubscribes a client from a room
func (h *Hub) Unsubscribe(client *Client, room string) {
	h.unsubscribe <- &UnsubscribeRequest{
		Client: client,
		Room:   room,
	}
}
