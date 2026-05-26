import { Bot, CheckCircle, Circle, Clock, ClipboardList, Database, GitBranch, ShieldAlert, Terminal, XCircle } from 'lucide-react'
import { useAgentStore } from '../store'
import { ToolCallData, ToolResultData } from '../types'
import clsx from 'clsx'

const capabilities = [
  { label: 'Filesystem + shell', icon: Terminal, state: 'online' },
  { label: 'Git tools', icon: GitBranch, state: 'online' },
  { label: 'Notion API', icon: Database, state: 'online' },
  { label: 'Approvals + risk', icon: ShieldAlert, state: 'online' },
  { label: 'Plans + steps', icon: ClipboardList, state: 'online' },
]

export default function ActivityPanel() {
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const events = useAgentStore((s) => activeThreadId ? (s.eventsByThread[activeThreadId] ?? []) : [])
  const isRunning = useAgentStore((s) => s.isRunning)
  const toolEvents = events.filter((ev) => ev.type === 'tool_call' || ev.type === 'tool_result' || ev.type === 'approval_required')
  const messages = events.filter((ev) => ev.type === 'message').length
  const toolCalls = events.filter((ev) => ev.type === 'tool_call').length
  const failures = events.filter((ev) => ev.type === 'tool_result' && !(ev.data as ToolResultData)?.success).length

  return (
    <aside className="hidden xl:flex w-[340px] flex-shrink-0 flex-col bg-notion-bg h-full">
      <div className="border-b border-notion-border px-5 py-4 drag-region">
        <div className="flex items-center justify-between no-drag">
          <div>
            <div className="text-[11px] uppercase tracking-[0.18em] text-notion-muted">Run monitor</div>
            <h2 className="mt-1 text-sm font-semibold text-notion-text">Atividade do agente</h2>
          </div>
          <div className={clsx('flex h-9 w-9 items-center justify-center rounded-notion border', isRunning ? 'border-notion-accent/40 bg-notion-selected text-notion-accent' : 'border-notion-border bg-notion-surface text-notion-muted')}>
            <Bot size={16} />
          </div>
        </div>
      </div>

      <div className="space-y-4 overflow-y-auto p-4">
        <section className="rounded-notion border border-notion-border bg-notion-surface p-4">
          <div className="mb-3 flex items-center gap-2 text-xs font-medium text-notion-text">
            <Clock size={14} className="text-notion-accent" />
            Run stats
          </div>
          <div className="grid grid-cols-3 gap-2">
            <Stat label="Tools" value={toolCalls} />
            <Stat label="Msgs" value={messages} />
            <Stat label="Fails" value={failures} tone={failures ? 'red' : 'green'} />
          </div>
        </section>

        <section className="rounded-notion border border-notion-border bg-notion-surface p-4">
          <div className="mb-3 text-xs font-medium text-notion-text">Tool timeline</div>
          {toolEvents.length === 0 ? (
            <p className="text-xs leading-5 text-notion-muted">As chamadas de ferramenta e aprovações aparecerão aqui, separadas da conversa principal.</p>
          ) : (
            <div className="space-y-2">
              {toolEvents.slice(-12).map((ev, idx) => {
                const call = ev.data as ToolCallData
                const result = ev.data as ToolResultData
                const success = result?.success
                return (
                  <div key={`${idx}-${ev.type}`} className="rounded-notion border border-notion-border bg-notion-bg p-3">
                    <div className="flex items-center gap-2 text-xs">
                      {ev.type === 'approval_required'
                        ? <ShieldAlert size={13} className="text-notion-orange" />
                        : ev.type === 'tool_result'
                          ? success ? <CheckCircle size={13} className="text-notion-green" /> : <XCircle size={13} className="text-notion-red" />
                          : <Terminal size={13} className="text-notion-accent" />}
                      <span className="font-mono text-notion-text">{call?.tool_name || result?.tool_name || (ev.data as any)?.action || 'tool'}</span>
                    </div>
                    <div className="mt-1 truncate text-[11px] text-notion-muted">{ev.type === 'tool_result' ? result?.output : ev.content || (ev.data as any)?.description}</div>
                  </div>
                )
              })}
            </div>
          )}
        </section>

        <section className="rounded-notion border border-notion-border bg-notion-surface p-4">
          <div className="mb-3 text-xs font-medium text-notion-text">Capacidades operacionais</div>
          <div className="space-y-2">
            {capabilities.map(({ label, icon: Icon, state }) => (
              <div key={label} className="flex items-center justify-between rounded-notion bg-notion-bg px-3 py-2">
                <div className="flex items-center gap-2 text-xs text-notion-muted">
                  <Icon size={13} />
                  {label}
                </div>
                <span className="flex items-center gap-1 text-[10px] uppercase tracking-[0.12em] text-notion-green">
                  <Circle size={6} className="fill-notion-green" />
                  {state}
                </span>
              </div>
            ))}
          </div>
        </section>
      </div>
    </aside>
  )
}

function Stat({ label, value, tone }: { label: string; value: number; tone?: 'green' | 'red' }) {
  return (
    <div className="rounded-notion border border-notion-border bg-notion-bg px-3 py-2">
      <div className={clsx('text-lg font-semibold', tone === 'green' ? 'text-notion-green' : tone === 'red' ? 'text-notion-red' : 'text-notion-text')}>{value}</div>
      <div className="text-[10px] uppercase tracking-[0.12em] text-notion-muted">{label}</div>
    </div>
  )
}
