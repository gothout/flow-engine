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
	IsEnd bool
	Seq   int
}

func (g *Goto) UnmarshalJSON(b []byte) error {
	// tentando com int
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		g.IsEnd = false
		g.Seq = n
		return nil
	}
	// tentando com string
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		s = strings.TrimSpace(s)
		if strings.EqualFold(s, "encerra") {
			g.IsEnd = true
			g.Seq = 0
			return nil
		}
		if num, err2 := strconv.Atoi(s); err2 == nil {
			g.IsEnd = false
			g.Seq = num
			return nil
		}
		return fmt.Errorf("goto invalido: %q", s)
	}
	return fmt.Errorf("goto invalido: %s", string(b))
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
	StartSeq int
	Steps    map[int]Step
}
