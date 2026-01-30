# Flow Engine

Projeto para executar fluxos de atendimento baseados em passos JSON (mensagem e opção).

## Exemplos de uso

### 1) Rodar o CLI com um fluxo de exemplo

O CLI lê `internal/flowcli/clinic.json` por padrão. Para rodar:

```bash
cd /workspace/flow-engine

go run ./internal/flowcli
```

### 2) Enviar eventos para o CLI

O CLI espera eventos JSON na entrada padrão. Exemplo de entrada:

```json
{"numero":"554799999999","protocolo":"TK123","mensagem":{"texto":"1"}}
```

Você pode enviar pela linha de comando assim:

```bash
echo '{"numero":"554799999999","protocolo":"TK123","mensagem":{"texto":"1"}}' | go run ./internal/flowcli
```

### 3) Exemplo de saída

A saída são mensagens JSON em cada passo do fluxo. Um exemplo simples:

```json
{"session_id":"554799999999|TK123","texto":"Olá seja bem vindo a clinica!"}
```

Quando o fluxo termina, o CLI envia `done`:

```json
{"session_id":"554799999999|TK123","done":true}
```

### 4) Estrutura mínima do fluxo

Um fluxo é uma lista de passos com `sequencia`, `tipo`, `mensagem` ou `opcao`, e `goto`:

```json
[
  {
    "sequencia": 1,
    "tipo": "mensagem",
    "mensagem": "Olá!",
    "goto": 2
  },
  {
    "sequencia": 2,
    "tipo": "opcao",
    "retorno": "$RETORNO",
    "opcoes": {
      "1,um": 3,
      "2,dois": 4
    },
    "goto": 5
  }
]
```
