import { motion } from 'framer-motion'
import { Terminal, CheckCircle, XCircle, Zap, AlertCircle } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { ChatEvent, ToolCallData, ToolResultData } from '../types'
import clsx from 'clsx'

const fadeInitial = { opacity: 0, y: 4 }
const fadeAnimate = { opacity: 1, y: 0 }
const fadeTransition = { duration: 0.12 }

interface Props {
  chatEvent: ChatEvent
}

export default function AgentStream({ chatEvent }: Props) {
  const { event } = chatEvent

  return (
    <motion.div
      initial={fadeInitial}
      animate={fadeAnimate}
      transition={fadeTransition}
      className="group"
    >
      {event.type === 'thinking' && (
        <div className="flex items-center gap-2 text-notion-muted text-xs py-0.5">
          <span className="animate-pulse">…</span>
          <span className="italic">{event.content}</span>
        </div>
      )}

      {event.type === 'message' && (
        <div className="bg-notion-panel rounded-notion px-4 py-3">
          <div className="prose-notion text-sm">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{event.content}</ReactMarkdown>
          </div>
        </div>
      )}

      {event.type === 'tool_call' && (
        <div className="flex items-start gap-2 py-1">
          <Terminal size={13} className="mt-0.5 text-notion-accent flex-shrink-0" />
          <div className="flex-1 min-w-0">
            <span className="text-xs font-mono text-notion-accent">
              {(event.data as ToolCallData)?.tool_name ?? 'tool'}
            </span>
            <span className="text-xs text-notion-muted ml-2">{event.content}</span>
          </div>
        </div>
      )}

      {event.type === 'tool_result' && (() => {
        const d = event.data as ToolResultData
        return (
          <div className={clsx(
            'rounded-notion px-3 py-2 text-xs font-mono flex items-start gap-2',
            d?.success ? 'bg-notion-green/10 text-notion-green' : 'bg-notion-red/10 text-notion-red'
          )}>
            {d?.success
              ? <CheckCircle size={12} className="mt-0.5 flex-shrink-0" />
              : <XCircle size={12} className="mt-0.5 flex-shrink-0" />}
            <span className="break-all line-clamp-4">{d?.output}</span>
          </div>
        )
      })()}

      {event.type === 'done' && (
        <div className="flex items-center gap-2 text-notion-green text-xs py-1">
          <Zap size={12} />
          <span>Concluído</span>
        </div>
      )}

      {event.type === 'error' && (
        <div className="flex items-start gap-2 bg-notion-red/10 rounded-notion px-3 py-2">
          <AlertCircle size={13} className="mt-0.5 flex-shrink-0 text-notion-red" />
          <span className="text-xs text-notion-red">{event.content}</span>
        </div>
      )}
    </motion.div>
  )
}
