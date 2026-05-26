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
  { label: 'Code', icon: Terminal, active: true },
  { label: 'Git', icon: GitBranch, active: false },
  { label: 'Notion', icon: Database, active: false },
]

export default function Sidebar() {
  const threads = useAgentStore((s) => s.threads)
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const setActiveThread = useAgentStore((s) => s.setActiveThread)
  const setShowSettings = useAgentStore((s) => s.setShowSettings)
  const { newThread, deleteThread } = useAgent()
  const [hovered, setHovered] = useState<string | null>(null)

  return (
    <aside className="w-[268px] flex-shrink-0 flex flex-col bg-codex-sidebar/95 h-full">
      <div className="px-4 pt-4 pb-3 drag-region">
        <div className="flex items-center justify-between no-drag">
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-xl bg-gradient-to-br from-codex-accent to-codex-accent2 flex items-center justify-center shadow-glow">
              <Bot size={16} className="text-white" />
            </div>
            <div>
              <div className="text-sm font-semibold tracking-tight text-codex-text">Notion Agent</div>
              <div className="flex items-center gap-1.5 text-[11px] text-codex-muted">
                <Circle size={7} className="fill-codex-green text-codex-green" />
                autonomous workspace
              </div>
            </div>
          </div>
          <button
            className="p-2 rounded-xl hover:bg-white/[0.06] text-codex-muted hover:text-codex-text transition-colors"
            onClick={newThread}
            title="Nova conversa"
          >
            <Plus size={16} />
          </button>
        </div>

        <button className="no-drag mt-4 flex items-center gap-2 w-full rounded-xl border border-codex-border bg-codex-card2/80 px-3 py-2 text-xs text-codex-muted hover:text-codex-text hover:border-codex-border2 transition-colors">
          <Search size={13} />
          Search threads, runs, files...
        </button>
      </div>

      <div className="px-3 py-2">
        <div className="text-[11px] uppercase tracking-[0.18em] text-codex-faint px-2 mb-2">Capabilities</div>
        <div className="grid grid-cols-3 gap-1.5">
          {capabilities.map(({ label, icon: Icon, active }) => (
            <div
              key={label}
              className={clsx(
                'rounded-xl border px-2 py-2 text-center text-[11px] transition-colors',
                active
                  ? 'border-codex-accent/40 bg-codex-accent/10 text-codex-text'
                  : 'border-codex-border bg-codex-card2/60 text-codex-faint'
              )}
            >
              <Icon size={13} className="mx-auto mb-1" />
              {label}
            </div>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-3 py-2">
        <div className="text-[11px] uppercase tracking-[0.18em] text-codex-faint px-2 mb-2">Threads</div>
        <AnimatePresence>
          {threads.length === 0 && (
            <div className="rounded-xl border border-dashed border-codex-border px-4 py-6 text-center text-xs text-codex-muted">
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
              className={clsx(
                'group flex items-center gap-2 px-2.5 py-2 rounded-xl cursor-pointer text-sm mb-1 transition-all border',
                activeThreadId === thread.id
                  ? 'bg-codex-card border-codex-border2 text-codex-text shadow-sm'
                  : 'border-transparent text-codex-muted hover:bg-white/[0.04] hover:text-codex-text'
              )}
              onClick={() => setActiveThread(thread.id)}
              onMouseEnter={() => setHovered(thread.id)}
              onMouseLeave={() => setHovered(null)}
            >
              <MessageSquare size={14} className="flex-shrink-0 opacity-70" />
              <div className="flex-1 min-w-0">
                <div className="truncate text-xs font-medium">{thread.title || 'Nova conversa'}</div>
                <div className="truncate text-[10px] text-codex-faint">{new Date(thread.updated_at).toLocaleDateString()}</div>
              </div>
              {hovered === thread.id && (
                <button
                  className="p-1 hover:text-codex-red rounded-lg transition"
                  onClick={(e) => { e.stopPropagation(); deleteThread(thread.id) }}
                >
                  <Trash2 size={12} />
                </button>
              )}
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      <div className="border-t border-codex-border px-3 py-3">
        <button
          onClick={() => setShowSettings(true)}
          className="flex items-center gap-2 px-3 py-2 w-full rounded-xl text-codex-muted hover:bg-white/[0.05] hover:text-codex-text transition-colors text-xs"
        >
          <Settings size={14} />
          Settings / provider
        </button>
      </div>
    </aside>
  )
}
