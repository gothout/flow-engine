package engine

import (
	"context"
	"fmt"
	"time"

	"flow-engine/internal/flow"
	"flow-engine/internal/session"
)

type Engine struct {
	Flow  *flow.Flow
	Store *session.Store
}

type IncomingEvent struct {
	SessionID string `json:"session_id,omitempty"`
	Numero    string `json:"numero,omitempty"`
	Protocolo string `json:"protocolo,omitempty"`
	Mensagem  struct {
		Texto string `json:"texto"`
	} `json:"mensagem"`
}

type OutMessage struct {
	SessionID string            `json:"session_id"`
	Texto     string            `json:"texto"`
	Done      bool              `json:"done,omitempty"`
	Vars      map[string]string `json:"vars,omitempty"`
	Error     string            `json:"error,omitempty"`
}

func New(f *flow.Flow, st *session.Store) *Engine {
	return &Engine{Flow: f, Store: st}
}

func resolveSessionID(in IncomingEvent) (string, error) {
	if in.SessionID != "" {
		return in.SessionID, nil
	}
	if in.Numero == "" || in.Protocolo == "" {
		return "", fmt.Errorf("session_id ausente e numero/protocolo incompletos")
	}
	return in.Numero + "|" + in.Protocolo, nil
}

func (e *Engine) HandleEventStream(
	ctx context.Context,
	in IncomingEvent,
	emit func(OutMessage) error,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if emit == nil {
		return fmt.Errorf("emit não pode ser nil")
	}

	sid, err := resolveSessionID(in)
	if err != nil {
		return err
	}

	sess := e.Store.GetOrCreate(sid, e.Flow.StartSeq)

	for {
		st, ok := e.Flow.Steps[sess.CurrentSeq]
		if !ok {
			return fmt.Errorf("sequencia %d não existe no fluxo", sess.CurrentSeq)
		}

		// somente mensagem por enquanto
		msg := OutMessage{
			SessionID: sid,
			Texto:     st.Mensagem,
		}

		// emite imediatamente
		if err := emit(msg); err != nil {
			return fmt.Errorf("emit falhou: %w", err)
		}

		// sleep antes de seguir para o goto
		if st.SleepMs > 0 {
			if err := sleep(ctx, time.Duration(st.SleepMs)*time.Millisecond); err != nil {
				return fmt.Errorf("sleep cancelado/erro: %w", err)
			}
		}

		if st.Goto.IsEnd {
			// emite “done” como mensagem separada (pra não depender de alterar a anterior)
			if err := emit(OutMessage{SessionID: sid, Done: true}); err != nil {
				return fmt.Errorf("emit done falhou: %w", err)
			}

			// reset pro início (seu comportamento atual)
			sess.CurrentSeq = e.Flow.StartSeq
			e.Store.Save(sess)
			return nil
		}

		sess.CurrentSeq = st.Goto.Seq
		e.Store.Save(sess)
	}
}

// Mantém seu HandleEvent antigo (compat), só que agora usa o stream por baixo
func (e *Engine) HandleEvent(ctx context.Context, in IncomingEvent) ([]OutMessage, error) {
	out := make([]OutMessage, 0, 4)
	err := e.HandleEventStream(ctx, in, func(m OutMessage) error {
		out = append(out, m)
		return nil
	})
	return out, err
}

// sleep para aguardar antes de seguir funcao
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
