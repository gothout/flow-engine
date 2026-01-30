package flow

import "encoding/json"

func Parse(dados []byte) (*Flow, error) {
	var passos []Step
	if err := json.Unmarshal(dados, &passos); err != nil {
		return nil, err
	}

	mapaPassos := make(map[int]Step, len(passos))
	sequenciaInicial := 0
	for _, passo := range passos {
		if sequenciaInicial == 0 || passo.Sequencia < sequenciaInicial {
			sequenciaInicial = passo.Sequencia
		}
		mapaPassos[passo.Sequencia] = passo
	}

	return &Flow{
		SequenciaInicial: sequenciaInicial,
		Passos:           mapaPassos,
	}, nil
}
