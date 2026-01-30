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

func (o Option) Execute(step flow.Step, sess *session.Session, input Input) (Result, error) {
	varName := normalizeVarName(step.Retorno)
	if varName == "" {
		return Result{}, fmt.Errorf("retorno vazio para opcao %d", step.Sequencia)
	}

	if sess.Vars == nil {
		sess.Vars = make(map[string]string)
	}
	delete(sess.Vars, varName)

	for optionKey, nextSeq := range step.Opcoes {
		if canonical, ok := matchOption(optionKey, input.UserText); ok {
			sess.Vars[varName] = canonical
			return Result{NextSeq: nextSeq}, nil
		}
	}

	// opcao invalida vai para o goto definido
	return Result{NextSeq: step.Goto.Seq}, nil
}
