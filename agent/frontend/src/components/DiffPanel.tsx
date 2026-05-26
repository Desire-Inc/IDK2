import { Terminal, Zap } from 'lucide-react'
import { useAgentStore } from '../store'
import { ToolCallData, ToolResultData } from '../types'

export default function DiffPanel() {
  const chatEvents = useAgentStore((s) => s.chatEvents)

  const toolEvents = chatEvents.filter(
    (e) => e.event.type === 'tool_call' || e.event.type === 'tool_result'
  )

  return (
    <div className="flex flex-col h-full bg-notion-panel">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-notion-border">
        <Terminal size={13} className="text-notion-accent" />
        <span className="text-xs font-semibold text-notion-muted uppercase tracking-widest">Atividade</span>
      </div>

      {/* Tool calls log */}
      <div className="flex-1 overflow-y-auto px-3 py-3 space-y-2">
        {toolEvents.length === 0 && (
          <p className="text-xs text-notion-muted text-center mt-8">
            As ações do agente aparecerão aqui.
          </p>
        )}
        {toolEvents.map((ce) => {
          const { event } = ce
          if (event.type === 'tool_call') {
            const d = event.data as ToolCallData
            return (
              <div key={ce.id} className="rounded-notion bg-notion-surface px-3 py-2">
                <div className="flex items-center gap-1.5 mb-1">
                  <Zap size={11} className="text-notion-accent" />
                  <span className="text-xs font-mono text-notion-accent font-medium">
                    {d?.tool_name}
                  </span>
                </div>
                <pre className="text-xs text-notion-muted overflow-x-auto whitespace-pre-wrap break-all">
                  {JSON.stringify(d?.arguments, null, 2)}
                </pre>
              </div>
            )
          }
          if (event.type === 'tool_result') {
            const d = event.data as ToolResultData
            return (
              <div
                key={ce.id}
                className={`rounded-notion px-3 py-2 text-xs font-mono break-all ${
                  d?.success
                    ? 'bg-notion-green/10 text-notion-green'
                    : 'bg-notion-red/10 text-notion-red'
                }`}
              >
                {d?.output?.slice(0, 400)}{(d?.output?.length ?? 0) > 400 ? '...' : ''}
              </div>
            )
          }
          return null
        })}
      </div>
    </div>
  )
}
