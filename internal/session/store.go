package session

import (
	"encoding/json"
	"sync"
	"time"
)

type Repository interface {
	GetOrCreate(idSessao string, sequenciaInicial int) *Session
	Save(sessao *Session)
}

type Encoder interface {
	Encode(sessao *Session) ([]byte, error)
}

type Session struct {
	IDSessao       string
	SequenciaAtual int
	Variaveis      map[string]string
	AtualizadoEm   time.Time
}

type Snapshot struct {
	ID             string            `json:"id"`
	SequenciaAtual int               `json:"current_seq"`
	Variaveis      map[string]string `json:"vars,omitempty"`
	AtualizadoEm   time.Time         `json:"last_updated"`
}

func (s *Session) Snapshot() Snapshot {
	return Snapshot{
		ID:             s.IDSessao,
		SequenciaAtual: s.SequenciaAtual,
		Variaveis:      s.Variaveis,
		AtualizadoEm:   s.AtualizadoEm,
	}
}

type JSONEncoder struct{}

func (JSONEncoder) Encode(sessao *Session) ([]byte, error) {
	return json.Marshal(sessao.Snapshot())
}

type Store struct {
	mu        sync.RWMutex
	dados     map[string]*Session
	tempoVida time.Duration
}

func NewStore(ttl time.Duration) *Store {
	return &Store{
		dados:     make(map[string]*Session),
		tempoVida: ttl,
	}
}

func (s *Store) GetOrCreate(idSessao string, sequenciaInicial int) *Session {
	agora := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	// limpeza simples
	if s.tempoVida > 0 {
		for chave, sessao := range s.dados {
			if agora.Sub(sessao.AtualizadoEm) > s.tempoVida {
				delete(s.dados, chave)
			}
		}
	}

	if sessao, ok := s.dados[idSessao]; ok {
		sessao.AtualizadoEm = agora
		return sessao
	}

	sessao := &Session{
		IDSessao:       idSessao,
		SequenciaAtual: sequenciaInicial,
		Variaveis:      make(map[string]string),
		AtualizadoEm:   agora,
	}
	s.dados[idSessao] = sessao
	return sessao
}

func (s *Store) Save(sessao *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessao.AtualizadoEm = time.Now()
	s.dados[sessao.IDSessao] = sessao
}

func (s *Store) Snapshot(idSessao string) (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessao, ok := s.dados[idSessao]
	if !ok {
		return Snapshot{}, false
	}
	return sessao.Snapshot(), true
}
