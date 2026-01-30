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

func (m Message) Execute(step flow.Step, sess *session.Session, input Input) (Result, error) {
	msg := renderTemplate(step.Mensagem, input, sess.Vars)
	return Result{
		Message: msg,
		NextSeq: step.Goto.Seq,
		Done:    step.Goto.IsEnd,
		Sleep:   time.Duration(step.SleepMs) * time.Millisecond,
	}, nil
}
