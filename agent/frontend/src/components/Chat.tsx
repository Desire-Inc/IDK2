import { useRef, useEffect } from 'react'
import { useAgentStore } from '../store'
import { useAgent } from '../hooks/useAgent'
import AgentStream from './AgentStream'
import Input from './Input'

export default function Chat() {
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const chatEvents = useAgentStore((s) => s.chatEvents)
  const status = useAgentStore((s) => s.status)
  const { sendMessage, stopAgent } = useAgent()
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [chatEvents])

  return (
    <div className="flex flex-col h-full bg-notion-surface">
      {/* Top bar */}
      <div className="flex items-center justify-between px-5 py-3 border-b border-notion-border drag-region">
        <span className="text-sm font-medium text-notion-text">
          {activeThreadId ? 'Conversa' : 'Notion Agent'}
        </span>
        <span
          className={`text-xs px-2 py-0.5 rounded-full font-medium ${
            status === 'running' ? 'bg-notion-accent/20 text-notion-accent' :
            status === 'waiting_approval' ? 'bg-notion-orange/20 text-notion-orange' :
            status === 'error' ? 'bg-notion-red/20 text-notion-red' :
            'text-notion-muted'
          }`}
        >
          {status === 'running' && '⏳ Executando'}
          {status === 'waiting_approval' && '⚠️ Aguardando aprovação'}
          {status === 'error' && '❌ Erro'}
          {status === 'idle' && (activeThreadId ? 'Pronto' : '')}
        </span>
      </div>

      {/* Messages / event stream */}
      <div className="flex-1 overflow-y-auto px-5 py-4 space-y-2">
        {chatEvents.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full gap-3 opacity-40">
            <div className="text-4xl">🤖</div>
            <p className="text-sm text-notion-muted text-center">
              Digite uma tarefa para o agente Notion.<br />
              Ex: “Crie uma database de tarefas com Status e Prazo”
            </p>
          </div>
        )}
        {chatEvents.map((ce) => (
          <AgentStream key={ce.id} chatEvent={ce} />
        ))}
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="px-4 pb-4 pt-2 border-t border-notion-border">
        <Input
          onSubmit={sendMessage}
          onStop={stopAgent}
          disabled={status === 'waiting_approval'}
          running={status === 'running'}
        />
      </div>
    </div>
  )
}
