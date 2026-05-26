import { motion } from 'framer-motion'
import { AlertCircle, Bot, CheckCircle, ChevronRight, ClipboardList, Loader2, ShieldAlert, Terminal, User, XCircle, Zap } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { ApprovalData, ChatEvent, ToolCallData, ToolResultData } from '../types'
import clsx from 'clsx'

const fadeIn = { opacity: 0, y: 6 }
const fadeVisible = { opacity: 1, y: 0 }
const fadeTrans = { duration: 0.14 }

interface Props {
  chatEvent: ChatEvent
}

function formatArgs(args: unknown) {
  if (!args) return ''
  if (typeof args === 'string') {
    try { return JSON.stringify(JSON.parse(args), null, 2) } catch { return args }
  }
  return JSON.stringify(args, null, 2)
}

function riskClass(risk?: string) {
  if (risk === 'dangerous') return 'text-notion-red border-notion-red/20 bg-notion-red/10'
  if (risk === 'medium') return 'text-notion-orange border-notion-orange/20 bg-notion-orange/10'
  return 'text-notion-green border-notion-green/20 bg-notion-green/10'
}

export default function AgentStream({ chatEvent }: Props) {
  const ev = chatEvent.event

  return (
    <motion.div initial={fadeIn} animate={fadeVisible} transition={fadeTrans}>
      {ev.type === 'user' && (
        <div className="flex justify-end">
          <div className="flex max-w-[82%] items-start gap-3">
            <div className="rounded-notion border border-notion-border bg-notion-panel px-4 py-3 text-sm leading-6 text-notion-text shadow-sm">
              {ev.content}
            </div>
            <div className="mt-1 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-notion-text text-notion-bg">
              <User size={14} />
            </div>
          </div>
        </div>
      )}

      {ev.type === 'plan' && (
        <div className="rounded-notion border border-notion-border bg-notion-surface px-4 py-3 text-sm text-notion-text">
          <div className="mb-1 flex items-center gap-2 text-xs uppercase tracking-[0.14em] text-notion-muted">
            <ClipboardList size={13} /> plano
          </div>
          {ev.content}
        </div>
      )}

      {ev.type === 'step' && (
        <div className="flex items-center gap-2 rounded-notion border border-notion-border bg-notion-surface px-3 py-2 text-xs text-notion-muted">
          <ChevronRight size={13} className="text-notion-accent" />
          <span>{ev.content}</span>
        </div>
      )}

      {ev.type === 'thinking' && (
        <div className="flex items-center gap-3 rounded-notion border border-notion-border bg-notion-surface px-4 py-3 text-xs text-notion-muted">
          <Loader2 size={14} className="animate-spin text-notion-accent" />
          <span>{ev.content || 'Pensando...'}</span>
        </div>
      )}

      {ev.type === 'reflection' && (
        <div className="rounded-notion border border-notion-border bg-notion-surface px-4 py-3 text-xs leading-5 text-notion-muted">
          {ev.content}
        </div>
      )}

      {ev.type === 'message' && (
        <div className="flex items-start gap-3">
          <div className="mt-1 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-notion-panel border border-notion-border text-notion-accent">
            <Bot size={14} />
          </div>
          <div className="min-w-0 flex-1 rounded-notion border border-notion-border bg-notion-surface px-5 py-4 text-sm text-notion-text">
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                p: (props) => <p className="mb-3 last:mb-0 leading-7" {...props} />,
                ul: (props) => <ul className="list-disc pl-5 mb-3 space-y-1.5" {...props} />,
                ol: (props) => <ol className="list-decimal pl-5 mb-3 space-y-1.5" {...props} />,
                li: (props) => <li className="text-notion-text leading-6" {...props} />,
                strong: (props) => <strong className="font-semibold text-white" {...props} />,
                a: (props) => <a className="text-notion-accent hover:underline" {...props} />,
                code: ({ children, className }) => {
                  const isBlock = Boolean(className?.includes('language-'))
                  return isBlock
                    ? <code className="block rounded-notion border border-notion-border bg-notion-bg p-4 text-xs font-mono overflow-x-auto my-3 text-notion-text">{children}</code>
                    : <code className="rounded border border-notion-border bg-notion-bg px-1.5 py-0.5 text-xs font-mono text-notion-text">{children}</code>
                },
              }}
            >
              {ev.content}
            </ReactMarkdown>
          </div>
        </div>
      )}

      {ev.type === 'approval_required' && (() => {
        const d = ev.data as ApprovalData
        return (
          <div className={clsx('rounded-notion border px-4 py-3 text-xs', riskClass(d?.risk))}>
            <div className="mb-1 flex items-center gap-2 font-medium">
              <ShieldAlert size={14} /> aprovação necessária · {d?.action}
            </div>
            <pre className="whitespace-pre-wrap break-words text-notion-muted">{d?.description}</pre>
          </div>
        )
      })()}

      {ev.type === 'tool_call' && (() => {
        const d = ev.data as ToolCallData
        return (
          <div className="rounded-notion border border-notion-border bg-notion-surface px-4 py-3">
            <div className="flex items-center gap-2 text-xs text-notion-muted">
              <Terminal size={14} className="text-notion-accent" />
              <span className="font-mono text-notion-accent">{d?.tool_name ?? 'tool'}</span>
              {d?.risk && <span className={clsx('rounded border px-1.5 py-0.5 text-[10px]', riskClass(d.risk))}>{d.risk}</span>}
              <ChevronRight size={13} />
              <span className="truncate">{ev.content}</span>
            </div>
            {d?.args && (
              <pre className="mt-2 max-h-32 overflow-auto rounded-notion bg-notion-bg p-3 text-[11px] text-notion-muted font-mono whitespace-pre-wrap">
                {formatArgs(d.args)}
              </pre>
            )}
          </div>
        )
      })()}

      {ev.type === 'tool_result' && (() => {
        const d = ev.data as ToolResultData
        return (
          <div className={clsx('rounded-notion border px-4 py-3 text-xs font-mono', d?.success ? 'border-notion-green/20 bg-notion-green/5 text-notion-green' : 'border-notion-red/20 bg-notion-red/5 text-notion-red')}>
            <div className="mb-2 flex items-center gap-2 font-sans text-[11px] uppercase tracking-[0.14em]">
              {d?.success ? <CheckCircle size={13} className="flex-shrink-0" /> : <XCircle size={13} className="flex-shrink-0" />}
              {d?.success ? 'tool result' : 'tool failed'} · {d?.tool_name}
            </div>
            <pre className="max-h-56 overflow-y-auto whitespace-pre-wrap break-words text-notion-muted">{d?.output}</pre>
          </div>
        )
      })()}

      {ev.type === 'cancelled' && (
        <div className="flex items-center gap-2 text-notion-orange text-xs py-1 pl-1">
          <XCircle size={12} />
          <span>{ev.content || 'Execução cancelada'}</span>
        </div>
      )}

      {ev.type === 'done' && (
        <div className="flex items-center gap-2 text-notion-green text-xs py-1 pl-1">
          <Zap size={12} />
          <span>Run concluída</span>
        </div>
      )}

      {ev.type === 'error' && (
        <div className="flex items-start gap-3 rounded-notion border border-notion-red/25 bg-notion-red/10 px-4 py-3">
          <AlertCircle size={15} className="mt-0.5 flex-shrink-0 text-notion-red" />
          <span className="text-xs leading-5 text-notion-red">{ev.content}</span>
        </div>
      )}
    </motion.div>
  )
}
