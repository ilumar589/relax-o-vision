package websocket

import (
	"encoding/json"
	"log/slog"
	"time"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

// Handler manages WebSocket connections
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// HandleConnection handles a new WebSocket connection
func (h *Handler) HandleConnection(c *fiberws.Conn) {
	client := &Client{
		ID:            uuid.New().String(),
		Conn:          c,
		Send:          make(chan *WSMessage, 256),
		Subscriptions: make(map[string]bool),
	}

	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// Start goroutines for reading and writing
	go h.writePump(client)
	h.readPump(client)
}

// readPump reads messages from the WebSocket connection
func (h *Handler) readPump(client *Client) {
	defer func() {
		h.hub.Unregister(client)
		client.Conn.Close()
	}()

	for {
		var msg WSMessage
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if fiberws.IsUnexpectedCloseError(err, fiberws.CloseGoingAway, fiberws.CloseAbnormalClosure) {
				slog.Error("WebSocket read error", "error", err)
			}
			break
		}

		// Handle subscription/unsubscription messages
		h.handleClientMessage(client, &msg)
	}
}

// writePump writes messages to the WebSocket connection
func (h *Handler) writePump(client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(fiberws.CloseMessage, []byte{})
				return
			}

			err := client.Conn.WriteJSON(message)
			if err != nil {
				slog.Error("WebSocket write error", "error", err)
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			if err := client.Conn.WriteMessage(fiberws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage handles messages received from clients
func (h *Handler) handleClientMessage(client *Client, msg *WSMessage) {
	switch msg.Type {
	case "subscribe":
		var subMsg SubscribeMessage
		if err := json.Unmarshal(msg.Payload, &subMsg); err != nil {
			slog.Error("Failed to unmarshal subscribe message", "error", err)
			return
		}
		h.hub.Subscribe(client, subMsg.Room)

	case "unsubscribe":
		var unsubMsg UnsubscribeMessage
		if err := json.Unmarshal(msg.Payload, &unsubMsg); err != nil {
			slog.Error("Failed to unmarshal unsubscribe message", "error", err)
			return
		}
		h.hub.Unsubscribe(client, unsubMsg.Room)

	default:
		slog.Warn("Unknown message type received", "type", msg.Type)
	}
}
