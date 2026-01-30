package flow

import "fmt"

func Validate(fluxo *Flow) error {
	if fluxo == nil || len(fluxo.Passos) == 0 {
		return fmt.Errorf("flow vazio")
	}
	if _, ok := fluxo.Passos[fluxo.SequenciaInicial]; !ok {
		return fmt.Errorf("start seq %d nao existe", fluxo.SequenciaInicial)
	}
	for sequencia, passo := range fluxo.Passos {
		if sequencia <= 0 {
			return fmt.Errorf("sequencia invalida: %d", sequencia)
		}
		if passo.SleepMs < 0 {
			return fmt.Errorf("step %d sleep_ms negativo", sequencia)
		}
		switch passo.Tipo {
		case StepMessage:
			if passo.Mensagem == "" {
				return fmt.Errorf("step %d (mensagem) sem campo mensagem", sequencia)
			}
			if !passo.Goto.Encerra {
				if _, ok := fluxo.Passos[passo.Goto.Sequencia]; !ok {
					return fmt.Errorf("step %d goto %d nao existe", sequencia, passo.Goto.Sequencia)
				}
			}
		case StepOption:
			if passo.Retorno == "" {
				return fmt.Errorf("step %d (opcao) sem campo retorno", sequencia)
			}
			if len(passo.Opcoes) == 0 {
				return fmt.Errorf("step %d (opcao) sem campo opcoes", sequencia)
			}
			if passo.Goto.Encerra {
				return fmt.Errorf("step %d (opcao) goto nao pode ser encerra", sequencia)
			}
			if _, ok := fluxo.Passos[passo.Goto.Sequencia]; !ok {
				return fmt.Errorf("step %d goto %d nao existe", sequencia, passo.Goto.Sequencia)
			}
			for _, destino := range passo.Opcoes {
				if _, ok := fluxo.Passos[destino]; !ok {
					return fmt.Errorf("step %d opcao com goto %d nao existe", sequencia, destino)
				}
			}
		default:
			return fmt.Errorf("step %d tipo nao suportado ainda: %s", sequencia, passo.Tipo)
		}
	}
	return nil
}
