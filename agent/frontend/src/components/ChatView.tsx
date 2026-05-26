import { useEffect, useMemo, useRef, useState } from 'react'
import { ArrowUp, Bot, Code2, Command, FileText, GitBranch, Loader2, Shield, Square, Terminal } from 'lucide-react'
import { useAgentStore } from '../store'
import { useAgent } from '../hooks/useAgent'
import AgentStream from './AgentStream'
import { ChatEvent } from '../types'
import clsx from 'clsx'

const suggestions = [
  { icon: Code2, label: 'Implementar feature', prompt: 'Analise o projeto e implemente a próxima melhoria mais importante.' },
  { icon: Terminal, label: 'Rodar diagnóstico', prompt: 'Leia a estrutura do projeto, rode os checks necessários e resuma o estado atual.' },
  { icon: GitBranch, label: 'Preparar commit', prompt: 'Revise as mudanças recentes, gere um resumo e prepare um commit limpo.' },
  { icon: FileText, label: 'Documentar', prompt: 'Atualize a documentação do projeto com arquitetura, setup e próximos passos.' },
]

export default function ChatView() {
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const activeThread = useAgentStore((s) => s.threads.find((t) => t.id === s.activeThreadId))
  const events = useAgentStore((s) =>
    activeThreadId ? (s.eventsByThread[activeThreadId] ?? []) : []
  )
  const isRunning = useAgentStore((s) => s.isRunning)
  const llmConfig = useAgentStore((s) => s.llmConfig)
  const { sendMessage } = useAgent()
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [events])

  const chatEvents: ChatEvent[] = useMemo(() => events.map((ev, i) => ({
    id: `${i}-${ev.type}`,
    event: ev,
  })), [events])

  const handleSend = (override?: string) => {
    const msg = (override ?? input).trim()
    if (!msg || isRunning) return
    setInput('')
    if (textareaRef.current) textareaRef.current.style.height = 'auto'
    sendMessage(msg)
  }

  const handleKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleInput = () => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 180)}px`
  }

  return (
    <div className="flex flex-col h-full bg-codex-bg/40">
      <header className="flex items-center justify-between px-6 py-4 border-b border-codex-border bg-codex-bg/70 backdrop-blur-xl flex-shrink-0 drag-region">
        <div className="no-drag min-w-0">
          <div className="flex items-center gap-2 text-[11px] uppercase tracking-[0.18em] text-codex-faint">
            <Command size={12} />
            active run
          </div>
          <h1 className="text-sm font-semibold text-codex-text truncate mt-1">
            {activeThread?.title || 'Nova conversa'}
          </h1>
        </div>
        <div className="no-drag flex items-center gap-2">
          <div className="hidden md:flex items-center gap-2 rounded-full border border-codex-border bg-codex-card2 px-3 py-1.5 text-xs text-codex-muted">
            <Bot size={13} />
            {llmConfig.model || 'mimo-v2.5-pro'}
          </div>
          <div className={clsx(
            'flex items-center gap-2 rounded-full border px-3 py-1.5 text-xs',
            isRunning
              ? 'border-codex-accent/40 bg-codex-accent/10 text-codex-text'
              : 'border-codex-border bg-codex-card2 text-codex-muted'
          )}>
            {isRunning ? <Loader2 size={13} className="animate-spin" /> : <Shield size={13} />}
            {isRunning ? 'executando' : 'idle'}
          </div>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto px-6 py-6">
        {chatEvents.length === 0 ? (
          <div className="mx-auto flex min-h-full max-w-3xl flex-col justify-center py-10">
            <div className="mb-7">
              <div className="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-codex-accent to-codex-accent2 shadow-glow">
                <Bot size={22} className="text-white" />
              </div>
              <h2 className="text-3xl font-semibold tracking-tight text-codex-text">O que vamos construir?</h2>
              <p className="mt-3 max-w-2xl text-sm leading-6 text-codex-muted">
                Um agente estilo Codex: conversa no centro, execução observável, ferramentas explícitas e contexto do projeto sempre visível.
              </p>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {suggestions.map(({ icon: Icon, label, prompt }) => (
                <button
                  key={label}
                  onClick={() => handleSend(prompt)}
                  className="group rounded-2xl border border-codex-border bg-codex-card2/70 p-4 text-left hover:border-codex-border2 hover:bg-codex-card transition-all"
                >
                  <Icon size={16} className="mb-3 text-codex-accent2" />
                  <div className="text-sm font-medium text-codex-text">{label}</div>
                  <div className="mt-1 text-xs leading-5 text-codex-muted group-hover:text-codex-text/80">{prompt}</div>
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div className="mx-auto max-w-4xl space-y-4">
            {chatEvents.map((ce) => (
              <AgentStream key={ce.id} chatEvent={ce} />
            ))}
            <div ref={bottomRef} />
          </div>
        )}
      </div>

      <div className="flex-shrink-0 px-6 pb-6 pt-2 bg-gradient-to-t from-codex-bg via-codex-bg/95 to-transparent">
        <div className="mx-auto max-w-4xl rounded-2xl border border-codex-border2 bg-codex-card/95 shadow-codex backdrop-blur-xl">
          <textarea
            ref={textareaRef}
            className="w-full bg-transparent text-sm text-codex-text placeholder-codex-faint resize-none outline-none min-h-[58px] max-h-[180px] px-4 pt-4 leading-6"
            placeholder="Peça uma mudança, investigação ou tarefa multi-step..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onInput={handleInput}
            onKeyDown={handleKey}
            rows={1}
            disabled={isRunning}
          />
          <div className="flex items-center justify-between gap-3 border-t border-codex-border px-3 py-2">
            <div className="flex items-center gap-2 text-[11px] text-codex-faint">
              <span className="rounded-md border border-codex-border px-1.5 py-0.5 font-mono">Enter</span>
              enviar
              <span className="rounded-md border border-codex-border px-1.5 py-0.5 font-mono">Shift Enter</span>
              nova linha
            </div>
            <button
              onClick={() => handleSend()}
              disabled={isRunning || !input.trim()}
              className={clsx(
                'flex items-center gap-2 rounded-xl px-3 py-2 text-xs font-medium transition-colors',
                isRunning
                  ? 'bg-codex-red/15 text-codex-red'
                  : input.trim()
                    ? 'bg-codex-text text-codex-bg hover:bg-white'
                    : 'bg-white/[0.05] text-codex-faint cursor-not-allowed'
              )}
            >
              {isRunning ? <Square size={13} /> : <ArrowUp size={13} />}
              {isRunning ? 'Stop' : 'Run'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
