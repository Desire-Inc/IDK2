import { useState } from 'react'
import { Plus, Trash2, MessageSquare, Settings, Bot, Search, Circle, GitBranch, Database, Terminal } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAgentStore } from '../store'
import { useAgent } from '../hooks/useAgent'
import clsx from 'clsx'

const itemInitial = { opacity: 0, x: -6 }
const itemAnimate = { opacity: 1, x: 0 }
const itemTransition = { duration: 0.15 }

const capabilities = [
  { label: 'Code', icon: Terminal },
  { label: 'Git', icon: GitBranch },
  { label: 'Notion', icon: Database },
]

export default function Sidebar() {
  const threads = useAgentStore((s) => s.threads)
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const setActiveThread = useAgentStore((s) => s.setActiveThread)
  const setShowSettings = useAgentStore((s) => s.setShowSettings)
  const { newThread, deleteThread } = useAgent()
  const [hovered, setHovered] = useState<string | null>(null)

  return (
    <aside className="w-[268px] flex-shrink-0 flex flex-col bg-notion-bg h-full">
      <div className="px-4 pt-4 pb-3 drag-region">
        <div className="flex items-center justify-between no-drag">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-notion bg-notion-panel border border-notion-border flex items-center justify-center">
              <Bot size={16} className="text-notion-accent" />
            </div>
            <div>
              <div className="text-sm font-semibold tracking-tight text-notion-text">Notion Agent</div>
              <div className="flex items-center gap-1.5 text-[11px] text-notion-muted">
                <Circle size={7} className="fill-notion-green text-notion-green" />
                autonomous workspace
              </div>
            </div>
          </div>
          <button className="p-2 rounded-notion hover:bg-notion-hover text-notion-muted hover:text-notion-text transition-colors" onClick={newThread} title="Nova conversa">
            <Plus size={16} />
          </button>
        </div>

        <button className="no-drag mt-4 flex items-center gap-2 w-full rounded-notion border border-notion-border bg-notion-surface px-3 py-2 text-xs text-notion-muted hover:text-notion-text hover:bg-notion-panel transition-colors">
          <Search size={13} />
          Buscar threads, runs, arquivos...
        </button>
      </div>

      <div className="px-3 py-2">
        <div className="text-[11px] uppercase tracking-[0.18em] text-notion-muted px-2 mb-2">Capabilities</div>
        <div className="grid grid-cols-3 gap-1.5">
          {capabilities.map(({ label, icon: Icon }) => (
            <div key={label} className="rounded-notion border border-notion-accent/30 bg-notion-selected px-2 py-2 text-center text-[11px] text-notion-text transition-colors">
              <Icon size={13} className="mx-auto mb-1 text-notion-accent" />
              {label}
            </div>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-3 py-2">
        <div className="text-[11px] uppercase tracking-[0.18em] text-notion-muted px-2 mb-2">Threads</div>
        <AnimatePresence>
          {threads.length === 0 && (
            <div className="rounded-notion border border-dashed border-notion-border px-4 py-6 text-center text-xs text-notion-muted">
              Nenhuma conversa ainda.<br />Crie uma run para começar.
            </div>
          )}
          {threads.map((thread) => (
            <motion.div
              key={thread.id}
              initial={itemInitial}
              animate={itemAnimate}
              exit={itemInitial}
              transition={itemTransition}
              className={clsx('group flex items-center gap-2 px-2.5 py-2 rounded-notion cursor-pointer text-sm mb-1 transition-all border', activeThreadId === thread.id ? 'bg-notion-panel border-notion-border text-notion-text shadow-sm' : 'border-transparent text-notion-muted hover:bg-notion-hover hover:text-notion-text')}
              onClick={() => setActiveThread(thread.id)}
              onMouseEnter={() => setHovered(thread.id)}
              onMouseLeave={() => setHovered(null)}
            >
              <MessageSquare size={14} className="flex-shrink-0 opacity-70" />
              <div className="flex-1 min-w-0">
                <div className="truncate text-xs font-medium">{thread.title || 'Nova conversa'}</div>
                <div className="truncate text-[10px] text-notion-muted">{new Date(thread.updated_at).toLocaleDateString()}</div>
              </div>
              {hovered === thread.id && (
                <button className="p-1 hover:text-notion-red rounded transition" onClick={(e) => { e.stopPropagation(); deleteThread(thread.id) }}>
                  <Trash2 size={12} />
                </button>
              )}
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      <div className="border-t border-notion-border px-3 py-3">
        <button onClick={() => setShowSettings(true)} className="flex items-center gap-2 px-3 py-2 w-full rounded-notion text-notion-muted hover:bg-notion-hover hover:text-notion-text transition-colors text-xs">
          <Settings size={14} />
          Settings / provider
        </button>
      </div>
    </aside>
  )
}
