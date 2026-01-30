package session

import (
	"sync"
	"time"
)

type Session struct {
	ID          string
	CurrentSeq  int
	Vars        map[string]string
	LastUpdated time.Time
}

type Store struct {
	mu   sync.RWMutex
	data map[string]*Session
	ttl  time.Duration
}

func NewStore(ttl time.Duration) *Store {
	return &Store{
		data: make(map[string]*Session),
		ttl:  ttl,
	}
}

func (s *Store) GetOrCreate(id string, startSeq int) *Session {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	// limpeza simples
	if s.ttl > 0 {
		for k, v := range s.data {
			if now.Sub(v.LastUpdated) > s.ttl {
				delete(s.data, k)
			}
		}
	}

	if sess, ok := s.data[id]; ok {
		sess.LastUpdated = now
		return sess
	}

	sess := &Session{
		ID:          id,
		CurrentSeq:  startSeq,
		Vars:        make(map[string]string),
		LastUpdated: now,
	}
	s.data[id] = sess
	return sess
}

func (s *Store) Save(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess.LastUpdated = time.Now()
	s.data[sess.ID] = sess
}
