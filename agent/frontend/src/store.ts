import { create } from 'zustand'
import { Thread, LLMConfig, AgentEventPayload } from './types'

interface AgentState {
  threads: Thread[]
  activeThreadId: string | null
  eventsByThread: Record<string, AgentEventPayload[]>
  isRunning: boolean
  pendingApproval: any | null
  showSettings: boolean
  llmConfig: LLMConfig

  setThreads: (threads: Thread[]) => void
  setActiveThread: (id: string | null) => void
  addEvent: (threadId: string, event: AgentEventPayload) => void
  clearEvents: (threadId: string) => void
  setRunning: (v: boolean) => void
  setPendingApproval: (v: any | null) => void
  setShowSettings: (v: boolean) => void
  setLLMConfig: (c: LLMConfig) => void
}

export const useAgentStore = create<AgentState>((set) => ({
  threads: [],
  activeThreadId: null,
  eventsByThread: {},
  isRunning: false,
  pendingApproval: null,
  showSettings: false,
  llmConfig: {
    provider: 'openai_compatible',
    model: 'mimo-v2.5-pro',
    base_url: 'https://opengateway.gitlawb.com/v1',
    api_key: '',
  },

  setThreads: (threads) => set({ threads }),

  setActiveThread: (id) => set({ activeThreadId: id, isRunning: false }),

  addEvent: (threadId, event) =>
    set((s) => ({
      eventsByThread: {
        ...s.eventsByThread,
        [threadId]: [...(s.eventsByThread[threadId] ?? []), event],
      },
    })),

  clearEvents: (threadId) =>
    set((s) => ({
      eventsByThread: { ...s.eventsByThread, [threadId]: [] },
    })),

  setRunning: (v) => set({ isRunning: v }),
  setPendingApproval: (v) => set({ pendingApproval: v }),
  setShowSettings: (v) => set({ showSettings: v }),
  setLLMConfig: (c) => set({ llmConfig: c }),
}))
