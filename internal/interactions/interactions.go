package interactions

import (
	"strings"
	"time"

	"flow-engine/internal/flow"
	"flow-engine/internal/session"
)

type Input struct {
	UserText string
}

type Result struct {
	Message string
	NextSeq int
	Done    bool
	Sleep   time.Duration
}

type Handler interface {
	Type() flow.StepType
	Execute(step flow.Step, sess *session.Session, input Input) (Result, error)
}

func normalizeVarName(name string) string {
	return strings.TrimPrefix(strings.TrimSpace(name), "$")
}

func renderTemplate(text string, input Input, vars map[string]string) string {
	out := text
	out = strings.ReplaceAll(out, "{{$USUARIO.TEXTO}}", input.UserText)
	for key, value := range vars {
		placeholder := "{{$" + key + "}}"
		out = strings.ReplaceAll(out, placeholder, value)
	}
	return out
}

func matchOption(optionsKey, userText string) (string, bool) {
	tokens := strings.Split(optionsKey, ",")
	if len(tokens) == 0 {
		return "", false
	}
	trimmed := make([]string, 0, len(tokens))
	for _, token := range tokens {
		val := strings.TrimSpace(token)
		if val != "" {
			trimmed = append(trimmed, val)
		}
	}
	for _, token := range trimmed {
		if strings.EqualFold(token, strings.TrimSpace(userText)) {
			return trimmed[len(trimmed)-1], true
		}
	}
	return "", false
}
