package store

import "github.com/qurt-quiz/qurt-quiz/internal/game"

// RoomStore is the persistence interface for rooms.
// Swap out the in-memory implementation for MongoDB without changing callers.
type RoomStore interface {
	Save(r *game.Room) error
	Get(id string) (*game.Room, error)
	Delete(id string) error
}
