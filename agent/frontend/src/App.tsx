import { useEffect } from 'react'
import { useAgentStore } from './store'
import { useAgent } from './hooks/useAgent'
import Sidebar from './components/Sidebar'
import Chat from './components/Chat'
import DiffPanel from './components/DiffPanel'
import ApprovalModal from './components/ApprovalModal'
import SettingsModal from './components/SettingsModal'

export default function App() {
  const { loadThreads, loadConfig } = useAgent()
  const pendingApproval = useAgentStore((s) => s.pendingApproval)
  const showSettings = useAgentStore((s) => s.showSettings)
  const chatEvents = useAgentStore((s) => s.chatEvents)

  useEffect(() => {
    loadThreads()
    loadConfig()
  }, [])

  // Show diff panel only when there are tool results with notion content
  const hasNotionActivity = chatEvents.some(
    (e) => e.event.type === 'tool_call' || e.event.type === 'tool_result'
  )

  return (
    <div className="flex h-screen w-screen bg-notion-bg text-notion-text font-sans overflow-hidden select-none">
      {/* Sidebar */}
      <Sidebar />

      {/* Main area */}
      <div className="flex flex-1 min-w-0">
        {/* Chat column */}
        <div className="flex flex-col flex-1 min-w-0 border-r border-notion-border">
          <Chat />
        </div>

        {/* Right panel — tools activity + diff */}
        {hasNotionActivity && (
          <div className="w-[380px] flex-shrink-0 hidden lg:flex flex-col">
            <DiffPanel />
          </div>
        )}
      </div>

      {/* Modals */}
      {pendingApproval && <ApprovalModal />}
      {showSettings && <SettingsModal />}
    </div>
  )
}
