package interactions

import (
	"strings"
	"time"

	"flow-engine/internal/flow"
	"flow-engine/internal/session"
)

type Entrada struct {
	TextoUsuario string
}

type Resultado struct {
	Mensagem   string
	ProxSeq    int
	Finalizado bool
	Espera     time.Duration
}

type Manipulador interface {
	Type() flow.StepType
	Execute(passo flow.Step, sessao *session.Session, entrada Entrada) (Resultado, error)
}

func normalizarNomeVariavel(nome string) string {
	return strings.TrimPrefix(strings.TrimSpace(nome), "$")
}

func renderizarTemplate(texto string, entrada Entrada, variaveis map[string]string) string {
	saida := texto
	saida = strings.ReplaceAll(saida, "{{$USUARIO.TEXTO}}", entrada.TextoUsuario)
	for chave, valor := range variaveis {
		marcador := "{{$" + chave + "}}"
		saida = strings.ReplaceAll(saida, marcador, valor)
	}
	return saida
}

func compararOpcao(chaveOpcoes, textoUsuario string) (string, bool) {
	partes := strings.Split(chaveOpcoes, ",")
	if len(partes) == 0 {
		return "", false
	}
	limpas := make([]string, 0, len(partes))
	for _, parte := range partes {
		valor := strings.TrimSpace(parte)
		if valor != "" {
			limpas = append(limpas, valor)
		}
	}
	for _, token := range limpas {
		if strings.EqualFold(token, strings.TrimSpace(textoUsuario)) {
			return limpas[len(limpas)-1], true
		}
	}
	return "", false
}
