package interactions

import (
	"fmt"

	"flow-engine/internal/flow"
	"flow-engine/internal/session"
)

type Option struct{}

func (o Option) Type() flow.StepType {
	return flow.StepOption
}

func (o Option) Execute(passo flow.Step, sessao *session.Session, entrada Entrada) (Resultado, error) {
	nomeVariavel := normalizarNomeVariavel(passo.Retorno)
	if nomeVariavel == "" {
		return Resultado{}, fmt.Errorf("retorno vazio para opcao %d", passo.Sequencia)
	}

	if sessao.Variaveis == nil {
		sessao.Variaveis = make(map[string]string)
	}
	delete(sessao.Variaveis, nomeVariavel)

	for chaveOpcao, proximaSeq := range passo.Opcoes {
		if canonico, ok := compararOpcao(chaveOpcao, entrada.TextoUsuario); ok {
			sessao.Variaveis[nomeVariavel] = canonico
			return Resultado{ProxSeq: proximaSeq}, nil
		}
	}

	// opcao invalida vai para o goto definido
	return Resultado{ProxSeq: passo.Goto.Sequencia}, nil
}
