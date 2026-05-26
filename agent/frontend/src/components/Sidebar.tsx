import { useState } from 'react'
import { Plus, Trash2, MessageSquare, Settings } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAgentStore } from '../store'
import { useAgent } from '../hooks/useAgent'
import clsx from 'clsx'

const itemInitial = { opacity: 0, x: -6 }
const itemAnimate = { opacity: 1, x: 0 }
const itemTransition = { duration: 0.15 }

export default function Sidebar() {
  const threads = useAgentStore((s) => s.threads)
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const setActiveThread = useAgentStore((s) => s.setActiveThread)
  const setShowSettings = useAgentStore((s) => s.setShowSettings)
  const { newThread, deleteThread } = useAgent()
  const [hovered, setHovered] = useState<string | null>(null)

  return (
    <aside className="w-[220px] flex-shrink-0 flex flex-col bg-notion-bg border-r border-notion-border h-full">
      <div className="flex items-center justify-between px-3 pt-4 pb-2">
        <span className="text-xs font-semibold uppercase tracking-widest text-notion-muted">
          Notion Agent
        </span>
        <button
          className="p-1 rounded hover:bg-notion-hover text-notion-muted hover:text-notion-text transition-colors"
          onClick={newThread}
          title="Nova conversa"
        >
          <Plus size={15} />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-1 py-1">
        <AnimatePresence>
          {threads.length === 0 && (
            <p className="text-center text-notion-muted text-xs mt-8 px-4">
              Nenhuma conversa ainda.
              <br />Clique em + para começar.
            </p>
          )}
          {threads.map((thread) => (
            <motion.div
              key={thread.id}
              initial={itemInitial}
              animate={itemAnimate}
              exit={itemInitial}
              transition={itemTransition}
              className={clsx(
                'group flex items-center gap-2 px-2.5 py-1.5 rounded-notion cursor-pointer text-sm mb-0.5 transition-colors',
                activeThreadId === thread.id
                  ? 'bg-notion-selected text-notion-text'
                  : 'text-notion-muted hover:bg-notion-hover hover:text-notion-text'
              )}
              onClick={() => setActiveThread(thread.id)}
              onMouseEnter={() => setHovered(thread.id)}
              onMouseLeave={() => setHovered(null)}
            >
              <MessageSquare size={13} className="flex-shrink-0 opacity-60" />
              <span className="flex-1 truncate text-xs">
                {thread.title || 'Nova conversa'}
              </span>
              {hovered === thread.id && (
                <button
                  className="p-0.5 hover:text-notion-red rounded transition"
                  onClick={(e) => { e.stopPropagation(); deleteThread(thread.id) }}
                >
                  <Trash2 size={11} />
                </button>
              )}
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      <div className="border-t border-notion-border px-2 py-2">
        <button
          onClick={() => setShowSettings(true)}
          className="flex items-center gap-2 px-2.5 py-1.5 w-full rounded-notion text-notion-muted hover:bg-notion-hover hover:text-notion-text transition-colors text-xs"
        >
          <Settings size={13} />
          Configurações
        </button>
      </div>
    </aside>
  )
}
