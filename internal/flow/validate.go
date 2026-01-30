package flow

import "fmt"

func Validate(f *Flow) error {
	if f == nil || len(f.Steps) == 0 {
		return fmt.Errorf("flow vazio")
	}
	if _, ok := f.Steps[f.StartSeq]; !ok {
		return fmt.Errorf("start seq %d nao existe", f.StartSeq)
	}
	for seq, st := range f.Steps {
		if seq <= 0 {
			return fmt.Errorf("sequencia invalida: %d", seq)
		}
		if st.SleepMs < 0 {
			return fmt.Errorf("step %d sleep_ms negativo", seq)
		}
		switch st.Tipo {
		case StepMessage:
			if st.Mensagem == "" {
				return fmt.Errorf("step %d (mensagem) sem campo mensagem", seq)
			}
			if !st.Goto.IsEnd {
				if _, ok := f.Steps[st.Goto.Seq]; !ok {
					return fmt.Errorf("step %d goto %d nao existe", seq, st.Goto.Seq)
				}
			}
		case StepOption:
			if st.Retorno == "" {
				return fmt.Errorf("step %d (opcao) sem campo retorno", seq)
			}
			if len(st.Opcoes) == 0 {
				return fmt.Errorf("step %d (opcao) sem campo opcoes", seq)
			}
			if st.Goto.IsEnd {
				return fmt.Errorf("step %d (opcao) goto nao pode ser encerra", seq)
			}
			if _, ok := f.Steps[st.Goto.Seq]; !ok {
				return fmt.Errorf("step %d goto %d nao existe", seq, st.Goto.Seq)
			}
			for _, dest := range st.Opcoes {
				if _, ok := f.Steps[dest]; !ok {
					return fmt.Errorf("step %d opcao com goto %d nao existe", seq, dest)
				}
			}
		default:
			return fmt.Errorf("step %d tipo nao suportado ainda: %s", seq, st.Tipo)
		}
	}
	return nil
}
