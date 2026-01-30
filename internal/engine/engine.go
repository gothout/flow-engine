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
	Fluxo         *flow.Flow
	Repositorio   session.Repository
	manipuladores map[flow.StepType]interactions.Manipulador
}

type IncomingEvent struct {
	IDSessao  string `json:"session_id,omitempty"`
	Numero    string `json:"numero,omitempty"`
	Protocolo string `json:"protocolo,omitempty"`
	Mensagem  struct {
		Texto string `json:"texto"`
	} `json:"mensagem"`
}

type OutMessage struct {
	IDSessao   string            `json:"session_id"`
	Texto      string            `json:"texto"`
	Finalizado bool              `json:"done,omitempty"`
	Variaveis  map[string]string `json:"vars,omitempty"`
	Erro       string            `json:"error,omitempty"`
}

func New(fluxo *flow.Flow, repositorio *session.Store) *Engine {
	eng := &Engine{
		Fluxo:       fluxo,
		Repositorio: repositorio,
	}
	eng.manipuladores = map[flow.StepType]interactions.Manipulador{
		flow.StepMessage: interactions.Message{},
		flow.StepOption:  interactions.Option{},
	}
	return eng
}

func resolverIDSessao(evento IncomingEvent) (string, error) {
	if evento.IDSessao != "" {
		return evento.IDSessao, nil
	}
	if evento.Numero == "" || evento.Protocolo == "" {
		return "", fmt.Errorf("session_id ausente e numero/protocolo incompletos")
	}
	return evento.Numero + "|" + evento.Protocolo, nil
}

func (e *Engine) HandleEventStream(
	contexto context.Context,
	evento IncomingEvent,
	emitir func(OutMessage) error,
) error {
	if contexto == nil {
		contexto = context.Background()
	}
	if emitir == nil {
		return fmt.Errorf("emit não pode ser nil")
	}

	idSessao, err := resolverIDSessao(evento)
	if err != nil {
		return err
	}

	sessao := e.Repositorio.GetOrCreate(idSessao, e.Fluxo.SequenciaInicial)
	entrada := interactions.Entrada{
		TextoUsuario: evento.Mensagem.Texto,
	}
	entradaConsumida := false

	for {
		passo, ok := e.Fluxo.Passos[sessao.SequenciaAtual]
		if !ok {
			return fmt.Errorf("sequencia %d não existe no fluxo", sessao.SequenciaAtual)
		}

		if passo.Tipo == flow.StepOption && entradaConsumida {
			e.Repositorio.Save(sessao)
			return nil
		}

		manipulador, ok := e.manipuladores[passo.Tipo]
		if !ok {
			return fmt.Errorf("tipo %s não registrado", passo.Tipo)
		}

		resultado, err := manipulador.Execute(passo, sessao, entrada)
		if err != nil {
			return err
		}

		if resultado.Mensagem != "" {
			msg := OutMessage{
				IDSessao:  idSessao,
				Texto:     resultado.Mensagem,
				Variaveis: sessao.Variaveis,
			}
			if err := emitir(msg); err != nil {
				return fmt.Errorf("emit falhou: %w", err)
			}
		}

		if resultado.Espera > 0 {
			if err := dormir(contexto, resultado.Espera); err != nil {
				return fmt.Errorf("sleep cancelado/erro: %w", err)
			}
		}

		if resultado.Finalizado {
			if err := emitir(OutMessage{IDSessao: idSessao, Finalizado: true}); err != nil {
				return fmt.Errorf("emit done falhou: %w", err)
			}

			sessao.SequenciaAtual = e.Fluxo.SequenciaInicial
			e.Repositorio.Save(sessao)
			return nil
		}

		pausarAposMensagem := entradaConsumida && passo.Tipo == flow.StepMessage

		if passo.Tipo == flow.StepOption {
			entradaConsumida = true
		}

		sessao.SequenciaAtual = resultado.ProxSeq
		e.Repositorio.Save(sessao)
		if pausarAposMensagem {
			return nil
		}
	}
}

// Mantém seu HandleEvent antigo (compat), só que agora usa o stream por baixo
func (e *Engine) HandleEvent(contexto context.Context, evento IncomingEvent) ([]OutMessage, error) {
	saida := make([]OutMessage, 0, 4)
	err := e.HandleEventStream(contexto, evento, func(m OutMessage) error {
		saida = append(saida, m)
		return nil
	})
	return saida, err
}

// dormir para aguardar antes de seguir funcao
func dormir(ctx context.Context, duracao time.Duration) error {
	if duracao <= 0 {
		return nil
	}
	temporizador := time.NewTimer(duracao)
	defer temporizador.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-temporizador.C:
		return nil
	}
}
