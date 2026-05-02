package hub

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/coder/websocket"
)

const sendBufSize = 64

// Client represents a single WebSocket connection.
// sessionID and roomID are written exactly once (by the room actor after join/reconnect)
// and read from the HTTP goroutine's readPump, so they're protected by a mutex.
type Client struct {
	mu        sync.RWMutex
	sessionID string
	roomID    string

	conn *websocket.Conn
	send chan []byte
	hub  *Hub
}

func newClient(conn *websocket.Conn, h *Hub) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, sendBufSize),
		hub:  h,
	}
}

func (c *Client) getSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

func (c *Client) getRoomID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.roomID
}

func (c *Client) setIdentity(sessionID, roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionID = sessionID
	c.roomID = roomID
}

// readPump runs in its own goroutine, forwarding WS messages to the hub/room.
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.unregister(c)
		c.conn.CloseNow()
	}()

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}

		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			c.sendError("invalid message format")
			continue
		}

		c.hub.dispatch(c, env)
	}
}

// writePump runs in its own goroutine, draining c.send to the WebSocket.
func (c *Client) writePump(ctx context.Context) {
	defer c.conn.CloseNow()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) sendError(msg string) {
	data, err := encode(MsgError, ErrorPayload{Message: msg})
	if err != nil {
		log.Printf("encode error: %v", err)
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) sendMsg(msgType string, payload any) {
	data, err := encode(msgType, payload)
	if err != nil {
		log.Printf("encode %s: %v", msgType, err)
		return
	}
	select {
	case c.send <- data:
	default:
		log.Printf("client send buffer full, dropping %s", msgType)
	}
}
