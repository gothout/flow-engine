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
	// Iniciar dessa maneira a sessao:
	// {"numero":"554799999999","protocolo":"TK123","mensagem":{"texto":"qualquer coisa"}}
	arquivoFluxo, err := os.ReadFile("./clinic.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro lendo clinic.json:", err)
		os.Exit(1)
	}

	fluxo, err := flow.Parse(arquivoFluxo)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro parse flow:", err)
		os.Exit(1)
	}
	if err := flow.Validate(fluxo); err != nil {
		fmt.Fprintln(os.Stderr, "flow inválido:", err)
		os.Exit(1)
	}

	repositorio := session.NewStore(30 * time.Minute)
	motor := engine.New(fluxo, repositorio)

	leitor := bufio.NewScanner(os.Stdin)
	leitor.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	codificador := json.NewEncoder(os.Stdout)
	codificador.SetEscapeHTML(false)

	for leitor.Scan() {
		linha := leitor.Text()
		if len(linha) == 0 {
			continue
		}

		var evento engine.IncomingEvent
		if err := json.Unmarshal([]byte(linha), &evento); err != nil {
			_ = codificador.Encode(engine.OutMessage{Erro: "json inválido: " + err.Error()})
			continue
		}

		err := motor.HandleEventStream(context.Background(), evento, func(m engine.OutMessage) error {
			return codificador.Encode(m) // imprime imediatamente
		})
		if err != nil {
			_ = codificador.Encode(engine.OutMessage{Erro: err.Error()})
			continue
		}
	}

	if err := leitor.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "erro lendo stdin:", err)
		os.Exit(1)
	}
}
