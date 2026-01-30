package interactions

import (
	"time"

	"flow-engine/internal/flow"
	"flow-engine/internal/session"
)

type Message struct{}

func (m Message) Type() flow.StepType {
	return flow.StepMessage
}

func (m Message) Execute(passo flow.Step, sessao *session.Session, entrada Entrada) (Resultado, error) {
	mensagem := renderizarTemplate(passo.Mensagem, entrada, sessao.Variaveis)
	return Resultado{
		Mensagem:   mensagem,
		ProxSeq:    passo.Goto.Sequencia,
		Finalizado: passo.Goto.Encerra,
		Espera:     time.Duration(passo.SleepMs) * time.Millisecond,
	}, nil
}
