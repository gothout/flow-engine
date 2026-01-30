package flow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type StepType string

const (
	StepMessage StepType = "mensagem"
	StepOption  StepType = "opcao"
)

type Goto struct {
	Encerra   bool
	Sequencia int
}

func (g *Goto) UnmarshalJSON(dados []byte) error {
	// tentando com int
	var numero int
	if err := json.Unmarshal(dados, &numero); err == nil {
		g.Encerra = false
		g.Sequencia = numero
		return nil
	}
	// tentando com string
	var texto string
	if err := json.Unmarshal(dados, &texto); err == nil {
		texto = strings.TrimSpace(texto)
		if strings.EqualFold(texto, "encerra") {
			g.Encerra = true
			g.Sequencia = 0
			return nil
		}
		if numero, err2 := strconv.Atoi(texto); err2 == nil {
			g.Encerra = false
			g.Sequencia = numero
			return nil
		}
		return fmt.Errorf("goto invalido: %q", texto)
	}
	return fmt.Errorf("goto invalido: %s", string(dados))
}

type Step struct {
	Sequencia  int      `json:"sequencia"`
	Tipo       StepType `json:"tipo"`
	Comentario string   `json:"comentario,omitempty"`
	// sleep opcional: espera ANTES de ir pro goto
	SleepMs int `json:"sleep_ms,omitempty"`

	// mensagem
	Mensagem string `json:"mensagem,omitempty"`

	// opcao
	Retorno string         `json:"retorno,omitempty"`
	Opcoes  map[string]int `json:"opcoes,omitempty"`
	Goto    Goto           `json:"goto"`
}

type Flow struct {
	SequenciaInicial int
	Passos           map[int]Step
}
