package session

import (
	"encoding/json"
	"sync"
	"time"
)

type Repository interface {
	GetOrCreate(id string, startSeq int) *Session
	Save(sess *Session)
}

type Encoder interface {
	Encode(sess *Session) ([]byte, error)
}

type Session struct {
	ID          string
	CurrentSeq  int
	Vars        map[string]string
	LastUpdated time.Time
}

type Snapshot struct {
	ID          string            `json:"id"`
	CurrentSeq  int               `json:"current_seq"`
	Vars        map[string]string `json:"vars,omitempty"`
	LastUpdated time.Time         `json:"last_updated"`
}

func (s *Session) Snapshot() Snapshot {
	return Snapshot{
		ID:          s.ID,
		CurrentSeq:  s.CurrentSeq,
		Vars:        s.Vars,
		LastUpdated: s.LastUpdated,
	}
}

type JSONEncoder struct{}

func (JSONEncoder) Encode(sess *Session) ([]byte, error) {
	return json.Marshal(sess.Snapshot())
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

func (s *Store) Snapshot(id string) (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.data[id]
	if !ok {
		return Snapshot{}, false
	}
	return sess.Snapshot(), true
}
