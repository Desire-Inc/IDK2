import { useEffect } from 'react'
import { useAgentStore } from './store'
import { useAgent } from './hooks/useAgent'
import Sidebar from './components/Sidebar'
import ChatView from './components/ChatView'
import ApprovalModal from './components/ApprovalModal'
import SettingsModal from './components/SettingsModal'
import ActivityPanel from './components/ActivityPanel'

export default function App() {
  const showSettings  = useAgentStore((s) => s.showSettings)
  const activeId      = useAgentStore((s) => s.activeThreadId)
  const { newThread } = useAgent()

  useEffect(() => {
    if (!activeId) newThread()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="h-screen w-screen overflow-hidden bg-codex-bg bg-codex-radial text-codex-text select-none">
      <div className="flex h-full w-full">
        <Sidebar />
        <main className="flex-1 min-w-0 overflow-hidden border-x border-codex-border/80">
          {activeId
            ? <ChatView />
            : <div className="flex items-center justify-center h-full"><p className="text-codex-muted text-sm">Inicializando workspace...</p></div>
          }
        </main>
        <ActivityPanel />
      </div>
      {showSettings && <SettingsModal />}
      <ApprovalModal />
    </div>
  )
}
