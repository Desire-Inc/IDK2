import { create } from 'zustand'
import { Thread, LLMConfig, AgentStatus, ChatEvent, ApprovalData } from '../types'

interface AgentStore {
  // Threads
  threads: Thread[]
  activeThreadId: string | null
  setThreads: (threads: Thread[]) => void
  setActiveThread: (id: string | null) => void
  addThread: (thread: Thread) => void
  removeThread: (id: string) => void
  updateThreadTitle: (id: string, title: string) => void

  // Chat events (live feed for current thread)
  chatEvents: ChatEvent[]
  addChatEvent: (event: ChatEvent) => void
  clearChatEvents: () => void

  // Agent status
  status: AgentStatus
  setStatus: (status: AgentStatus) => void

  // Pending approval
  pendingApproval: ApprovalData | null
  setPendingApproval: (data: ApprovalData | null) => void

  // LLM config
  llmConfig: LLMConfig
  setLLMConfig: (config: LLMConfig) => void

  // Settings modal
  showSettings: boolean
  setShowSettings: (show: boolean) => void
}

export const useAgentStore = create<AgentStore>((set) => ({
  threads: [],
  activeThreadId: null,
  setThreads: (threads) => set({ threads }),
  setActiveThread: (id) => set({ activeThreadId: id, chatEvents: [] }),
  addThread: (thread) => set((s) => ({ threads: [thread, ...s.threads] })),
  removeThread: (id) => set((s) => ({ threads: s.threads.filter((t) => t.id !== id) })),
  updateThreadTitle: (id, title) =>
    set((s) => ({
      threads: s.threads.map((t) => (t.id === id ? { ...t, title } : t)),
    })),

  chatEvents: [],
  addChatEvent: (event) => set((s) => ({ chatEvents: [...s.chatEvents, event] })),
  clearChatEvents: () => set({ chatEvents: [] }),

  status: 'idle',
  setStatus: (status) => set({ status }),

  pendingApproval: null,
  setPendingApproval: (data) => set({ pendingApproval: data }),

  llmConfig: { provider: 'anthropic', model: 'claude-sonnet-4-5' },
  setLLMConfig: (config) => set({ llmConfig: config }),

  showSettings: false,
  setShowSettings: (show) => set({ showSettings: show }),
}))
