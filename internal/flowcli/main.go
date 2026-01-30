package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"flow-engine/internal/engine"
	"flow-engine/internal/flow"
	"flow-engine/internal/session"
)

func main() {
	// Iniciar dessa maneira a sessao
	// {"numero":"554799999999","protocolo":"TK123","mensagem":{"texto":"qualquer coisa"}}
	flowBytes, err := os.ReadFile("./clinic.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro lendo clinic.json:", err)
		os.Exit(1)
	}

	f, err := flow.Parse(flowBytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro parse flow:", err)
		os.Exit(1)
	}
	if err := flow.Validate(f); err != nil {
		fmt.Fprintln(os.Stderr, "flow inválido:", err)
		os.Exit(1)
	}

	store := session.NewStore(30 * time.Minute)
	eng := engine.New(f, store)

	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	for sc.Scan() {
		line := sc.Text()
		if len(line) == 0 {
			continue
		}

		var in engine.IncomingEvent
		if err := json.Unmarshal([]byte(line), &in); err != nil {
			_ = enc.Encode(engine.OutMessage{Error: "json inválido: " + err.Error()})
			continue
		}

		err := eng.HandleEventStream(context.Background(), in, func(m engine.OutMessage) error {
			return enc.Encode(m) // imprime imediatamente
		})
		if err != nil {
			_ = enc.Encode(engine.OutMessage{Error: err.Error()})
			continue
		}
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "erro lendo stdin:", err)
		os.Exit(1)
	}
}
