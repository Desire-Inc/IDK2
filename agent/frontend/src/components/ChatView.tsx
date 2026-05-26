import { useEffect, useRef } from 'react'
import { Send } from 'lucide-react'
import { useAgentStore } from '../store'
import { useAgent } from '../hooks/useAgent'
import AgentStream from './AgentStream'
import { useState } from 'react'

export default function ChatView() {
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const events = useAgentStore((s) =>
    activeThreadId ? (s.eventsByThread[activeThreadId] ?? []) : []
  )
  const isRunning = useAgentStore((s) => s.isRunning)
  const { sendMessage } = useAgent()
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom on new events
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [events])

  const handleSend = () => {
    const msg = input.trim()
    if (!msg || isRunning) return
    setInput('')
    sendMessage(msg)
  }

  const handleKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex flex-col h-full bg-notion-bg">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-notion-border flex-shrink-0">
        <h1 className="text-sm font-semibold text-notion-text">Conversa</h1>
        {isRunning && (
          <span className="text-xs text-notion-accent animate-pulse">Executando...</span>
        )}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-6 py-4 space-y-3">
        {events.length === 0 && (
          <div className="flex items-center justify-center h-full">
            <p className="text-notion-muted text-sm text-center">
              Digite uma tarefa para o agente executar.
            </p>
          </div>
        )}
        {events.map((ev, i) => (
          <AgentStream key={i} chatEvent= id: String(i), event: ev  />
        ))}
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="flex-shrink-0 border-t border-notion-border px-4 py-3">
        <div className="flex items-end gap-2 bg-notion-panel rounded-notion border border-notion-border px-3 py-2">
          <textarea
            className="flex-1 bg-transparent text-sm text-notion-text placeholder-notion-muted resize-none outline-none min-h-[20px] max-h-[120px]"
            placeholder="Digite uma tarefa para o agente..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKey}
            rows={1}
            disabled={isRunning}
          />
          <button
            onClick={handleSend}
            disabled={isRunning || !input.trim()}
            className="p-1.5 rounded-notion bg-notion-accent text-white disabled:opacity-40 hover:bg-notion-accent/80 transition-colors flex-shrink-0"
          >
            <Send size={14} />
          </button>
        </div>
      </div>
    </div>
  )
}
