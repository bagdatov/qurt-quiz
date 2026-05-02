package memory

import (
	"fmt"
	"sync"

	"github.com/qurt-quiz/qurt-quiz/internal/game"
)

type Store struct {
	mu    sync.RWMutex
	rooms map[string]*game.Room
}

func New() *Store {
	return &Store{rooms: make(map[string]*game.Room)}
}

func (s *Store) Save(r *game.Room) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[r.ID] = r
	return nil
}

func (s *Store) Get(id string) (*game.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[id]
	if !ok {
		return nil, fmt.Errorf("room %q not found", id)
	}
	return r, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rooms, id)
	return nil
}
