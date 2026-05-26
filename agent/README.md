# Notion Agent

Um agente de IA no estilo Codex/Claude que controla seu workspace Notion de forma autônoma.

## Stack

| Camada | Tecnologia |
|--------|------------|
| Desktop | [Wails v2](https://wails.io) (Go + WebView) |
| Frontend | React 18 + TypeScript + TailwindCSS |
| Backend | Go 1.22 |
| LLMs | Anthropic Claude, OpenAI GPT, Ollama (local) |
| DB local | SQLite (memória de threads) |
| Execução de código | Docker sandbox / fallback local |

## Estrutura

```
agent/
├── main.go               # Entry point Wails
├── wails.json            # Config da janela desktop
├── go.mod
├── internal/
│   ├── app/              # Wails App, ReAct loop, memória SQLite, eventos
│   ├── llm/              # Provider interface + Anthropic, OpenAI, Ollama
│   └── tools/            # Registry + ferramentas (Notion, código, git, filesystem)
├── cmd/                  # CLI (notion-agent run "tarefa")
└── frontend/             # React UI
    ├── src/
    │   ├── components/   # Sidebar, Chat, DiffPanel, AgentStream, Modals, Input
    │   ├── store/        # Zustand store global
    │   ├── hooks/        # useAgent (Wails bridge)
    │   └── types/        # TypeScript types
    └── index.html
```

## Ferramentas disponíveis para o agente

### Notion
| Tool | Descrição |
|------|------------|
| `notion_search` | Busca páginas e databases |
| `notion_page_view` | Lê uma página por ID |
| `notion_page_create` | Cria uma nova página |
| `notion_page_delete` | Arquiva uma página |
| `notion_page_set_properties` | Atualiza propriedades |
| `notion_db_list` | Lista databases |
| `notion_db_query` | Consulta linhas de uma database |
| `notion_db_create` | Cria uma nova database |
| `notion_db_add_row` | Adiciona uma linha |
| `notion_block_list` | Lista blocos de uma página |
| `notion_block_append` | Adiciona blocos |
| `notion_user_me` | Info do usuário autenticado |
| `notion_comment_list` | Lista comentários |

### Código
| Tool | Descrição |
|------|------------|
| `run_code` | Executa Python, JavaScript ou Bash em sandbox Docker (fallback local) |

### Git
| Tool | Descrição |
|------|------------|
| `git_status` | Status do repositório |
| `git_diff` | Diff staged/unstaged |
| `git_commit` | Commit com mensagem |
| `git_push` | Push para remote |
| `git_log` | Histórico de commits |

### Filesystem
| Tool | Descrição |
|------|------------|
| `read_file` | Lê um arquivo local |
| `write_file` | Escreve um arquivo (cria diretórios) |
| `list_directory` | Lista arquivos e pastas |

## Como rodar

### Pré-requisitos

```bash
# Go 1.22+
brew install go

# Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Node.js 20+
brew install node
```

### Configuração

```bash
cp .env.example .env
# edite .env com suas chaves
```

```env
NOTION_TOKEN=secret_xxxx          # Integration token do Notion
ANTHROPIC_API_KEY=sk-ant-xxxx     # Chave da Anthropic (opcional)
OPENAI_API_KEY=sk-xxxx            # Chave da OpenAI (opcional)
```

### Desenvolvimento (app desktop)

```bash
cd agent
wails dev
```

### Build de produção

```bash
cd agent
wails build
# executável em build/bin/notion-agent
```

### CLI (sem interface gráfica)

```bash
cd agent
go run ./cmd/... run "Crie uma database de tarefas com Status e Prazo"

# Com OpenAI
go run ./cmd/... run --provider openai --model gpt-4o "Resuma meu workspace"

# Com Ollama local
go run ./cmd/... run --provider ollama --model llama3 "Liste todas as páginas"
```

## Design

3 colunas no estilo Codex / Claude:

```
┏━━━━━━━━━━━━━━┓ ┏━━━━━━━━━━━━━━━━━━━━━━━━┓ ┏━━━━━━━━━━━━━━━━━┓
┃ Threads     ┃ ┃ Chat / Streaming    ┃ ┃ Atividade (tools) ┃
┃ (histórico) ┃ ┃                     ┃ ┃ + diff Notion     ┃
┃             ┃ ┃ [input]             ┃ ┃                   ┃
┗━━━━━━━━━━━━━━┛ ┗━━━━━━━━━━━━━━━━━━━━━━━━┛ ┗━━━━━━━━━━━━━━━━━┛
```

Cores Notion dark: `#191919` sidebar • `#1E1E1E` chat • `#252525` painel • `#2383E2` accent

## Fluxo ReAct

```
User input
    ↓
 System prompt (PT-BR) + histórico
    ↓
 LLM (stream)
    ├─ thinking → UI: "..."
    ├─ text     → UI: message bubble
    └─ tool_call → executor → tool_result → next iteration
    ↓
 done / max_iterations
```
