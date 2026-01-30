package engine

import (
	"context"
	"fmt"
	"time"

	"flow-engine/internal/flow"
	"flow-engine/internal/interactions"
	"flow-engine/internal/session"
)

type Engine struct {
	Flow     *flow.Flow
	Store    session.Repository
	handlers map[flow.StepType]interactions.Handler
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
	eng := &Engine{
		Flow:  f,
		Store: st,
	}
	eng.handlers = map[flow.StepType]interactions.Handler{
		flow.StepMessage: interactions.Message{},
		flow.StepOption:  interactions.Option{},
	}
	return eng
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
	input := interactions.Input{
		UserText: in.Mensagem.Texto,
	}
	inputConsumed := false

	for {
		st, ok := e.Flow.Steps[sess.CurrentSeq]
		if !ok {
			return fmt.Errorf("sequencia %d não existe no fluxo", sess.CurrentSeq)
		}

		if st.Tipo == flow.StepOption && inputConsumed {
			e.Store.Save(sess)
			return nil
		}

		handler, ok := e.handlers[st.Tipo]
		if !ok {
			return fmt.Errorf("tipo %s não registrado", st.Tipo)
		}

		result, err := handler.Execute(st, sess, input)
		if err != nil {
			return err
		}

		if result.Message != "" {
			msg := OutMessage{
				SessionID: sid,
				Texto:     result.Message,
				Vars:      sess.Vars,
			}
			if err := emit(msg); err != nil {
				return fmt.Errorf("emit falhou: %w", err)
			}
		}

		if result.Sleep > 0 {
			if err := sleep(ctx, result.Sleep); err != nil {
				return fmt.Errorf("sleep cancelado/erro: %w", err)
			}
		}

		if result.Done {
			if err := emit(OutMessage{SessionID: sid, Done: true}); err != nil {
				return fmt.Errorf("emit done falhou: %w", err)
			}

			sess.CurrentSeq = e.Flow.StartSeq
			e.Store.Save(sess)
			return nil
		}

		if st.Tipo == flow.StepOption {
			inputConsumed = true
		}

		sess.CurrentSeq = result.NextSeq
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
