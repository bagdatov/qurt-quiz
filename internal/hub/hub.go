package hub

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/qurt-quiz/qurt-quiz/internal/game"
	"github.com/qurt-quiz/qurt-quiz/internal/pack"
	"github.com/qurt-quiz/qurt-quiz/internal/store"
)

const defaultAnswerDuration = 30 * time.Second

// Hub manages all active rooms and routes incoming WebSocket connections.
type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]*Room
	store   store.RoomStore
	packDir string
}

func New(s store.RoomStore, packDir string) *Hub {
	return &Hub{
		rooms:   make(map[string]*Room),
		store:   s,
		packDir: packDir,
	}
}

// ServeWS upgrades an HTTP request to WebSocket and starts the client pumps.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // tighten to specific origins in production
	})
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}

	c := newClient(conn, h)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.writePump(ctx)
	c.readPump(ctx)
}

// dispatch routes an envelope from a client to the correct handler.
func (h *Hub) dispatch(c *Client, env Envelope) {
	switch env.Type {
	case MsgCreateRoom:
		h.handleCreateRoom(c, env.Payload)
	default:
		h.routeToRoom(c, env)
	}
}

// routeToRoom forwards an envelope to the client's current room actor.
// For ENTER_ROOM, the room ID is extracted from the payload.
func (h *Hub) routeToRoom(c *Client, env Envelope) {
	roomID := c.getRoomID()

	if roomID == "" && env.Type == MsgEnterRoom {
		var p EnterRoomPayload
		if err := json.Unmarshal(env.Payload, &p); err == nil {
			roomID = p.RoomID
		}
	}

	if roomID == "" {
		c.sendError("not in a room — send ENTER_ROOM or CREATE_ROOM first")
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		c.sendError("room not found")
		return
	}
	room.send(roomMsg{client: c, env: env})
}

// unregister notifies the client's room that the connection closed.
func (h *Hub) unregister(c *Client) {
	roomID := c.getRoomID()
	if roomID == "" {
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return
	}
	room.send(roomMsg{client: c, internal: &internalMsg{typ: MsgInternalClientLeft}})
}

// removeRoom is called by the room actor when it shuts down.
func (h *Hub) removeRoom(id string) {
	h.mu.Lock()
	delete(h.rooms, id)
	h.mu.Unlock()
}

// handleCreateRoom creates a room and makes the connecting client the host.
// All state mutations (including the initial broadcastState) happen via the room's
// incoming channel so they run inside the actor goroutine, not the HTTP goroutine.
func (h *Hub) handleCreateRoom(c *Client, raw json.RawMessage) {
	var p CreateRoomPayload
	if err := json.Unmarshal(raw, &p); err != nil || p.SessionID == "" || p.Name == "" || p.PackID == "" {
		c.sendError("CREATE_ROOM requires session_id, name, pack_id, lang")
		return
	}

	loaded, err := pack.Load(h.packDir, p.PackID, p.Lang)
	if err != nil {
		c.sendError("failed to load pack: " + err.Error())
		return
	}

	roomID := shortID()
	gr := &game.Room{
		ID:     roomID,
		PackID: p.PackID,
		Lang:   p.Lang,
		Players: map[string]*game.Player{
			p.SessionID: {
				ID:        p.SessionID,
				Name:      p.Name,
				IsHost:    true,
				Connected: true,
			},
		},
		State: game.GameState{
			Phase:          game.PhaseLobby,
			Pack:           loaded,
			UsedCells:      make(map[string]bool),
			AnswerDuration: defaultAnswerDuration,
		},
	}

	if err := h.store.Save(gr); err != nil {
		c.sendError("internal error creating room")
		return
	}

	// Set identity on the client before the room actor starts, so the actor sees
	// it when processing the _INIT message. This write happens in the HTTP goroutine
	// before go room.run() is scheduled, satisfying the happens-before requirement.
	c.setIdentity(p.SessionID, roomID)

	room := newRoom(gr, h.store, h)
	room.clients[p.SessionID] = c

	h.mu.Lock()
	h.rooms[roomID] = room
	h.mu.Unlock()

	// Queue the initial broadcast as an actor message so broadcastState() runs
	// inside the room goroutine, not the HTTP goroutine. This avoids a data race
	// on room.clients between the two goroutines.
	room.incoming <- roomMsg{internal: &internalMsg{typ: MsgInternalInit}}
	go room.run()
}

// shortID generates a 6-character uppercase room code.
func shortID() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
