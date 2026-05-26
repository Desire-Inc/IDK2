import { Bot, CheckCircle, Circle, Clock, Database, FileText, GitBranch, Terminal, XCircle } from 'lucide-react'
import { useAgentStore } from '../store'
import { ToolCallData, ToolResultData } from '../types'
import clsx from 'clsx'

const roadmap = [
  { label: 'Filesystem + shell', icon: Terminal, state: 'online' },
  { label: 'Git tools', icon: GitBranch, state: 'next' },
  { label: 'Notion tools', icon: Database, state: 'next' },
  { label: 'Run summaries', icon: FileText, state: 'planned' },
]

export default function ActivityPanel() {
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const events = useAgentStore((s) => activeThreadId ? (s.eventsByThread[activeThreadId] ?? []) : [])
  const isRunning = useAgentStore((s) => s.isRunning)
  const toolEvents = events.filter((ev) => ev.type === 'tool_call' || ev.type === 'tool_result')
  const messages = events.filter((ev) => ev.type === 'message').length
  const toolCalls = events.filter((ev) => ev.type === 'tool_call').length
  const failures = events.filter((ev) => ev.type === 'tool_result' && !(ev.data as ToolResultData)?.success).length

  return (
    <aside className="hidden xl:flex w-[340px] flex-shrink-0 flex-col bg-codex-sidebar/80 h-full">
      <div className="border-b border-codex-border px-5 py-4 drag-region">
        <div className="flex items-center justify-between no-drag">
          <div>
            <div className="text-[11px] uppercase tracking-[0.18em] text-codex-faint">Run monitor</div>
            <h2 className="mt-1 text-sm font-semibold text-codex-text">Atividade do agente</h2>
          </div>
          <div className={clsx(
            'flex h-9 w-9 items-center justify-center rounded-xl border',
            isRunning ? 'border-codex-accent/40 bg-codex-accent/10 text-codex-accent' : 'border-codex-border bg-codex-card2 text-codex-muted'
          )}>
            <Bot size={16} />
          </div>
        </div>
      </div>

      <div className="space-y-4 overflow-y-auto p-4">
        <section className="rounded-2xl border border-codex-border bg-codex-card2/70 p-4">
          <div className="mb-3 flex items-center gap-2 text-xs font-medium text-codex-text">
            <Clock size={14} className="text-codex-accent2" />
            Run stats
          </div>
          <div className="grid grid-cols-3 gap-2">
            <Stat label="Tools" value={toolCalls} />
            <Stat label="Msgs" value={messages} />
            <Stat label="Fails" value={failures} tone={failures ? 'red' : 'green'} />
          </div>
        </section>

        <section className="rounded-2xl border border-codex-border bg-codex-card2/70 p-4">
          <div className="mb-3 text-xs font-medium text-codex-text">Tool timeline</div>
          {toolEvents.length === 0 ? (
            <p className="text-xs leading-5 text-codex-muted">As chamadas de ferramenta aparecerão aqui, separadas da conversa principal.</p>
          ) : (
            <div className="space-y-2">
              {toolEvents.slice(-10).map((ev, idx) => {
                const call = ev.data as ToolCallData
                const result = ev.data as ToolResultData
                const success = result?.success
                return (
                  <div key={`${idx}-${ev.type}`} className="rounded-xl border border-codex-border bg-codex-bg/40 p-3">
                    <div className="flex items-center gap-2 text-xs">
                      {ev.type === 'tool_result'
                        ? success ? <CheckCircle size={13} className="text-codex-green" /> : <XCircle size={13} className="text-codex-red" />
                        : <Terminal size={13} className="text-codex-accent2" />}
                      <span className="font-mono text-codex-text">{call?.tool_name || result?.tool_name || 'tool'}</span>
                    </div>
                    <div className="mt-1 truncate text-[11px] text-codex-faint">
                      {ev.type === 'tool_result' ? result?.output : ev.content}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </section>

        <section className="rounded-2xl border border-codex-border bg-codex-card2/70 p-4">
          <div className="mb-3 text-xs font-medium text-codex-text">Roadmap operacional</div>
          <div className="space-y-2">
            {roadmap.map(({ label, icon: Icon, state }) => (
              <div key={label} className="flex items-center justify-between rounded-xl bg-codex-bg/35 px-3 py-2">
                <div className="flex items-center gap-2 text-xs text-codex-muted">
                  <Icon size={13} />
                  {label}
                </div>
                <span className={clsx(
                  'flex items-center gap-1 text-[10px] uppercase tracking-[0.12em]',
                  state === 'online' ? 'text-codex-green' : state === 'next' ? 'text-codex-amber' : 'text-codex-faint'
                )}>
                  <Circle size={6} className={state === 'online' ? 'fill-codex-green' : ''} />
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
    <div className="rounded-xl border border-codex-border bg-codex-bg/40 px-3 py-2">
      <div className={clsx(
        'text-lg font-semibold',
        tone === 'green' ? 'text-codex-green' : tone === 'red' ? 'text-codex-red' : 'text-codex-text'
      )}>{value}</div>
      <div className="text-[10px] uppercase tracking-[0.12em] text-codex-faint">{label}</div>
    </div>
  )
}
