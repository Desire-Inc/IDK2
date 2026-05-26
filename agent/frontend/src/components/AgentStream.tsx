import { motion } from 'framer-motion'
import { AlertCircle, Bot, CheckCircle, ChevronRight, Loader2, Terminal, User, XCircle, Zap } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { ChatEvent, ToolCallData, ToolResultData } from '../types'
import clsx from 'clsx'

const fadeIn = { opacity: 0, y: 6 }
const fadeVisible = { opacity: 1, y: 0 }
const fadeTrans = { duration: 0.14 }

interface Props {
  chatEvent: ChatEvent
}

export default function AgentStream({ chatEvent }: Props) {
  const ev = chatEvent.event

  return (
    <motion.div initial={fadeIn} animate={fadeVisible} transition={fadeTrans}>
      {ev.type === 'user' && (
        <div className="flex justify-end">
          <div className="flex max-w-[82%] items-start gap-3">
            <div className="rounded-2xl border border-codex-border2 bg-codex-card px-4 py-3 text-sm leading-6 text-codex-text shadow-sm">
              {ev.content}
            </div>
            <div className="mt-1 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-codex-text text-codex-bg">
              <User size={14} />
            </div>
          </div>
        </div>
      )}

      {ev.type === 'thinking' && (
        <div className="flex items-center gap-3 rounded-2xl border border-codex-border bg-codex-card2/60 px-4 py-3 text-xs text-codex-muted">
          <Loader2 size={14} className="animate-spin text-codex-accent" />
          <span>{ev.content || 'Pensando...'}</span>
        </div>
      )}

      {ev.type === 'message' && (
        <div className="flex items-start gap-3">
          <div className="mt-1 flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-codex-accent to-codex-accent2 text-white shadow-glow">
            <Bot size={14} />
          </div>
          <div className="min-w-0 flex-1 rounded-2xl border border-codex-border bg-codex-card2/80 px-5 py-4 text-sm text-codex-text">
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                p: (props) => <p className="mb-3 last:mb-0 leading-7" {...props} />,
                ul: (props) => <ul className="list-disc pl-5 mb-3 space-y-1.5" {...props} />,
                ol: (props) => <ol className="list-decimal pl-5 mb-3 space-y-1.5" {...props} />,
                li: (props) => <li className="text-codex-text leading-6" {...props} />,
                strong: (props) => <strong className="font-semibold text-white" {...props} />,
                a: (props) => <a className="text-codex-blue hover:underline" {...props} />,
                code: ({ children, className }) => {
                  const isBlock = Boolean(className?.includes('language-'))
                  return isBlock
                    ? <code className="block rounded-xl border border-codex-border bg-codex-bg p-4 text-xs font-mono overflow-x-auto my-3 text-codex-text">{children}</code>
                    : <code className="rounded-md border border-codex-border bg-codex-bg px-1.5 py-0.5 text-xs font-mono text-codex-text">{children}</code>
                },
              }}
            >
              {ev.content}
            </ReactMarkdown>
          </div>
        </div>
      )}

      {ev.type === 'tool_call' && (() => {
        const d = ev.data as ToolCallData
        return (
          <div className="rounded-2xl border border-codex-border bg-codex-card2/60 px-4 py-3">
            <div className="flex items-center gap-2 text-xs text-codex-muted">
              <Terminal size={14} className="text-codex-accent2" />
              <span className="font-mono text-codex-accent2">{d?.tool_name ?? 'tool'}</span>
              <ChevronRight size={13} />
              <span className="truncate">{ev.content}</span>
            </div>
            {d?.args && (
              <pre className="mt-2 max-h-24 overflow-auto rounded-xl bg-codex-bg/80 p-3 text-[11px] text-codex-faint font-mono whitespace-pre-wrap">
                {JSON.stringify(d.args, null, 2)}
              </pre>
            )}
          </div>
        )
      })()}

      {ev.type === 'tool_result' && (() => {
        const d = ev.data as ToolResultData
        return (
          <div className={clsx(
            'rounded-2xl border px-4 py-3 text-xs font-mono',
            d?.success
              ? 'border-codex-green/20 bg-codex-green/5 text-codex-green'
              : 'border-codex-red/20 bg-codex-red/5 text-codex-red'
          )}>
            <div className="mb-2 flex items-center gap-2 font-sans text-[11px] uppercase tracking-[0.14em]">
              {d?.success
                ? <CheckCircle size={13} className="flex-shrink-0" />
                : <XCircle size={13} className="flex-shrink-0" />}
              {d?.success ? 'tool result' : 'tool failed'} · {d?.tool_name}
            </div>
            <pre className="max-h-56 overflow-y-auto whitespace-pre-wrap break-words text-codex-muted">{d?.output}</pre>
          </div>
        )
      })()}

      {ev.type === 'done' && (
        <div className="flex items-center gap-2 text-codex-green text-xs py-1 pl-1">
          <Zap size={12} />
          <span>Run concluída</span>
        </div>
      )}

      {ev.type === 'error' && (
        <div className="flex items-start gap-3 rounded-2xl border border-codex-red/25 bg-codex-red/10 px-4 py-3">
          <AlertCircle size={15} className="mt-0.5 flex-shrink-0 text-codex-red" />
          <span className="text-xs leading-5 text-codex-red">{ev.content}</span>
        </div>
      )}
    </motion.div>
  )
}
