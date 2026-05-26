import { motion } from 'framer-motion'
import { Terminal, CheckCircle, XCircle, Zap, AlertCircle, User } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { ChatEvent, ToolCallData, ToolResultData } from '../types'
import clsx from 'clsx'

const fadeIn = { opacity: 0, y: 4 }
const fadeVisible = { opacity: 1, y: 0 }
const fadeTrans = { duration: 0.12 }

interface Props {
  chatEvent: ChatEvent
}

export default function AgentStream({ chatEvent }: Props) {
  const ev = chatEvent.event

  return (
    <motion.div initial={fadeIn} animate={fadeVisible} transition={fadeTrans}>

      {ev.type === 'user' && (
        <div className="flex justify-end mb-1">
          <div className="flex items-start gap-2 max-w-[80%]">
            <div className="bg-notion-accent/15 rounded-notion px-4 py-2.5 text-sm text-notion-text">
              {ev.content}
            </div>
            <User size={15} className="mt-1 text-notion-muted flex-shrink-0" />
          </div>
        </div>
      )}

      {ev.type === 'thinking' && (
        <div className="flex items-center gap-2 text-notion-muted text-xs py-0.5 pl-1">
          <span className="animate-pulse text-notion-accent">●</span>
          <span className="italic">{ev.content}</span>
        </div>
      )}

      {ev.type === 'message' && (
        <div className="bg-notion-panel rounded-notion px-4 py-3 text-sm text-notion-text">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              p: (props) => <p className="mb-2 last:mb-0 leading-relaxed" {...props} />,
              ul: (props) => <ul className="list-disc pl-4 mb-2 space-y-1" {...props} />,
              ol: (props) => <ol className="list-decimal pl-4 mb-2 space-y-1" {...props} />,
              li: (props) => <li className="text-notion-text" {...props} />,
              strong: (props) => <strong className="font-semibold" {...props} />,
              code: ({ children, className }) => {
                const isBlock = Boolean(className?.includes('language-'))
                return isBlock
                  ? <code className="block bg-notion-bg rounded p-3 text-xs font-mono overflow-x-auto my-2">{children}</code>
                  : <code className="bg-notion-bg px-1 py-0.5 rounded text-xs font-mono">{children}</code>
              },
            }}
          >
            {ev.content}
          </ReactMarkdown>
        </div>
      )}

      {ev.type === 'tool_call' && (
        <div className="flex items-center gap-2 py-0.5 pl-1">
          <Terminal size={12} className="text-notion-accent flex-shrink-0" />
          <span className="text-xs font-mono text-notion-accent">
            {(ev.data as ToolCallData)?.tool_name ?? 'tool'}
          </span>
          <span className="text-xs text-notion-muted truncate">{ev.content}</span>
        </div>
      )}

      {ev.type === 'tool_result' && (() => {
        const d = ev.data as ToolResultData
        return (
          <div className={clsx(
            'rounded-notion px-3 py-2 text-xs font-mono flex items-start gap-2 max-h-40 overflow-y-auto',
            d?.success ? 'bg-notion-green/10 text-notion-green' : 'bg-notion-red/10 text-notion-red'
          )}>
            {d?.success
              ? <CheckCircle size={12} className="mt-0.5 flex-shrink-0" />
              : <XCircle size={12} className="mt-0.5 flex-shrink-0" />}
            <pre className="whitespace-pre-wrap break-all">{d?.output}</pre>
          </div>
        )
      })()}

      {ev.type === 'done' && (
        <div className="flex items-center gap-1.5 text-notion-green text-xs py-0.5 pl-1">
          <Zap size={11} />
          <span>Concluído</span>
        </div>
      )}

      {ev.type === 'error' && (
        <div className="flex items-start gap-2 bg-notion-red/10 rounded-notion px-3 py-2">
          <AlertCircle size={13} className="mt-0.5 flex-shrink-0 text-notion-red" />
          <span className="text-xs text-notion-red">{ev.content}</span>
        </div>
      )}
    </motion.div>
  )
}
