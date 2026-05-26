import { useEffect } from 'react'
import { useAgentStore } from './store'
import { useAgent } from './hooks/useAgent'
import Sidebar from './components/Sidebar'
import ChatView from './components/ChatView'
import ApprovalModal from './components/ApprovalModal'
import SettingsModal from './components/SettingsModal'

export default function App() {
  const showSettings  = useAgentStore((s) => s.showSettings)
  const activeId      = useAgentStore((s) => s.activeThreadId)
  const { newThread } = useAgent()

  useEffect(() => {
    if (!activeId) newThread()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-notion-bg text-notion-text select-none">
      <Sidebar />
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {activeId
          ? <ChatView />
          : <div className="flex items-center justify-center h-full"><p className="text-notion-muted text-sm">Iniciando...</p></div>
        }
      </main>
      {showSettings && <SettingsModal />}
      <ApprovalModal />
    </div>
  )
}
