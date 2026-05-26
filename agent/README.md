# Notion Agent

> AI agent that controls your Notion workspace — just like Codex or Claude, but Notion-native.

---

## O que é

O **Notion Agent** é um agente ReAct (Raciocinar + Agir) que recebe uma tarefa em linguagem natural e executa ações no seu workspace Notion de forma autônoma:

- Criar e editar páginas, databases e blocos
- Consultar e filtrar dados
- Executar código (Python, JS, Bash) em sandbox
- Operar com Git/filesystem
- Pedir aprovação antes de ações destrutivas

A interface é uma aplicação desktop (Wails + React), com design inspirado no Codex/Claude mas com a identidade visual do Notion.

---

## Stack

| Camada | Tecnologia |
|---|---|
| Backend | Go 1.22, [Wails v2](https://wails.io) |
| Frontend | React 18 + TypeScript + TailwindCSS |
| LLM | Anthropic Claude / OpenAI GPT / Ollama (local) |
| Memória | SQLite (via `go-sqlite3`) |
| Notion | Notion REST API v1 |
| Sandbox | Docker (fallback: local) |

---

## Estrutura

```
agent/
├── main.go                  # Entry point Wails
├── wails.json               # Configuração Wails
├── go.mod
├── Makefile
├── internal/
│   ├── app/
│   │   ├── app.go           # Wails App (métodos expostos ao frontend)
│   │   ├── loop.go          # ReAct loop principal
│   │   ├── memory.go        # SQLite: threads e mensagens
│   │   └── events.go        # Tipos de evento
│   ├── llm/
│   │   ├── provider.go      # Interface Provider + tipos
│   │   ├── anthropic.go     # Claude (SSE streaming)
│   │   ├── openai.go        # GPT (go-openai)
│   │   └── ollama.go        # Ollama (local)
│   └── tools/
│       ├── registry.go      # Registry de tools
│       ├── notion.go        # 12 tools Notion
│       ├── code.go          # Sandbox Python/JS/Bash
│       ├── git.go           # Git tools
│       └── filesystem.go    # Read/write filesystem
├── cmd/
│   ├── root.go              # Cobra root
│   └── agent.go             # CLI: notion-agent run "tarefa"
└── frontend/
    ├── index.html
    ├── package.json
    ├── vite.config.ts
    ├── tailwind.config.js
    └── src/
        ├── App.tsx
        ├── main.tsx
        ├── types/index.ts
        ├── store/index.ts
        ├── hooks/useAgent.ts
        ├── styles/globals.css
        └── components/
            ├── Sidebar.tsx
            ├── Chat.tsx
            ├── AgentStream.tsx
            ├── Input.tsx
            ├── DiffPanel.tsx
            ├── ApprovalModal.tsx
            └── SettingsModal.tsx
```

---

## Setup

### Pré-requisitos

- Go 1.22+
- Node.js 20+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- (Opcional) Docker — para sandbox de código

### Variáveis de ambiente

```bash
# Para Anthropic (padrão)
export ANTHROPIC_API_KEY=sk-ant-...

# Para OpenAI
export OPENAI_API_KEY=sk-...

# Notion Integration Token
export NOTION_TOKEN=secret_...
```

> **Como obter o NOTION_TOKEN:** vá em [notion.so/my-integrations](https://www.notion.so/my-integrations), crie uma integração, copie o token e adicione a integração às páginas/databases que o agente deve acessar.

### Rodar em modo desenvolvimento

```bash
cd agent
make dev
# ou:
wails dev
```

### Build desktop (produção)

```bash
cd agent
make build
# Binário gerado em: build/bin/notion-agent
```

### CLI (sem interface gráfica)

```bash
cd agent
go run . run "Crie uma database chamada Sprint com colunas Tarefa, Status e Responsável"

# Com provider específico
go run . run --provider openai --model gpt-4o "Resuma todas as páginas da pasta Projetos"
```

---

## Como funciona

```
Usuário digita tarefa
        ↓
  Loop ReAct (Go)
        ↓
  LLM raciocina → gera tool_call
        ↓
  Tool executada (Notion API / código / git)
        ↓
  Resultado devolvido ao LLM
        ↓
  LLM decide: continuar ou responder
        ↓
  Ações perigosas → ApprovalModal (usuário aprova/recusa)
        ↓
  Resposta final exibida no chat
```

### Tools disponíveis

| Tool | Descrição |
|---|---|
| `notion_search` | Busca páginas e databases |
| `notion_page_view` | Lê uma página por ID |
| `notion_page_create` | Cria página |
| `notion_page_delete` | Arquiva página |
| `notion_page_set_properties` | Atualiza propriedades |
| `notion_db_list` | Lista databases |
| `notion_db_query` | Consulta rows com filtros |
| `notion_db_create` | Cria database |
| `notion_db_add_row` | Adiciona row a database |
| `notion_block_list` | Lista blocos de uma página |
| `notion_block_append` | Adiciona blocos |
| `notion_user_me` | Info do usuário/bot autenticado |
| `notion_comment_list` | Lista comentários |
| `run_code` | Executa Python/JS/Bash em sandbox |
| `read_file` | Lê arquivo local |
| `write_file` | Escreve arquivo local |
| `list_directory` | Lista diretório |
| `git_status` | `git status` |
| `git_diff` | `git diff` |
| `git_commit` | Stage + commit |
| `git_push` | Push para remote |
| `git_log` | Log de commits |

---

## Configurações (UI)

Clique em **Configurações** na sidebar:

- **Provedor:** Anthropic / OpenAI / Ollama
- **Modelo:** claude-sonnet-4-5, gpt-4o, llama3, etc.
- **API Key:** salva localmente via Wails preferences
- **URL Ollama:** para instâncias locais customizadas

---

## Roadmap

- [ ] Phase 1 ✅ — Estrutura core (LLM + tools + UI + ReAct loop)
- [ ] Phase 2 — Notion OAuth (login com conta do usuário)
- [ ] Phase 3 — Melhorias de UX: histórico, diff visual de blocos, multi-agent

---

## Licença

MIT — Desire Inc
