import { motion, AnimatePresence } from 'framer-motion'
import { AlertTriangle, Check, X } from 'lucide-react'
import { useAgentStore } from '../store'
import { useAgent } from '../hooks/useAgent'

const overlayInitial = { opacity: 0 }
const overlayAnimate = { opacity: 1 }
const modalInitial = { opacity: 0, scale: 0.96, y: 8 }
const modalAnimate = { opacity: 1, scale: 1, y: 0 }
const modalTransition = { duration: 0.18, ease: 'easeOut' }

export default function ApprovalModal() {
  const pendingApproval = useAgentStore((s) => s.pendingApproval)
  const { approveAction } = useAgent()

  if (!pendingApproval) return null

  return (
    <AnimatePresence>
      <motion.div
        initial={overlayInitial}
        animate={overlayAnimate}
        exit={overlayInitial}
        className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50"
      >
        <motion.div
          initial={modalInitial}
          animate={modalAnimate}
          exit={modalInitial}
          transition={modalTransition}
          className="bg-notion-surface rounded-notion shadow-2xl border border-notion-border w-[440px] p-6"
        >
          <div className="flex items-center gap-3 mb-4">
            <div className="p-2 bg-notion-orange/20 rounded-notion">
              <AlertTriangle size={18} className="text-notion-orange" />
            </div>
            <div>
              <h3 className="font-semibold text-sm text-notion-text">Ação requer aprovação</h3>
              <p className="text-xs text-notion-muted">{pendingApproval.action}</p>
            </div>
          </div>

          <p className="text-sm text-notion-muted mb-5 leading-relaxed">
            {pendingApproval.description}
          </p>

          <div className="flex gap-2 justify-end">
            <button
              onClick={() => approveAction(false)}
              className="flex items-center gap-1.5 px-4 py-2 rounded-notion text-sm text-notion-muted hover:bg-notion-hover transition-colors"
            >
              <X size={14} />
              Recusar
            </button>
            <button
              onClick={() => approveAction(true)}
              className="flex items-center gap-1.5 px-4 py-2 rounded-notion text-sm bg-notion-accent text-white hover:bg-notion-accent/80 transition-colors"
            >
              <Check size={14} />
              Aprovar
            </button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  )
}
