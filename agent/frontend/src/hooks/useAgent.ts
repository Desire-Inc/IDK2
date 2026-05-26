import { useEffect, useCallback } from 'react'
import { useAgentStore } from '../store'
import { AgentEvent, ApprovalData, ChatEvent } from '../types'

// Wails runtime bridge
declare const window: Window & {
  go: {
    app: {
      App: {
        ListThreads: () => Promise<import('../types').Thread[]>
        GetMessages: (threadID: string) => Promise<import('../types').Message[]>
        NewThread: (title: string) => Promise<string>
        DeleteThread: (threadID: string) => Promise<void>
        RunAgent: (threadID: string, input: string) => Promise<void>
        StopAgent: () => Promise<void>
        ApproveAction: (approved: boolean) => Promise<void>
        GetLLMConfig: () => Promise<import('../types').LLMConfig>
        SetLLMConfig: (config: import('../types').LLMConfig) => Promise<void>
        GetAvailableModels: () => Promise<Record<string, string[]>>
      }
    }
  }
  runtime: {
    EventsOn: (event: string, callback: (data: unknown) => void) => (() => void)
    EventsOff: (event: string) => void
  }
}

const wails = () => window.go.app.App

export function useAgent() {
  const store = useAgentStore()

  // Subscribe to agent events from Go backend
  useEffect(() => {
    const offEvent = window.runtime.EventsOn('agent:event', (data: unknown) => {
      const event = data as AgentEvent
      const chatEvent: ChatEvent = {
        id: `${Date.now()}-${Math.random()}`,
        event,
      }
      store.addChatEvent(chatEvent)

      if (event.type === 'done' || event.type === 'error') {
        store.setStatus('idle')
      } else if (event.type === 'thinking' || event.type === 'tool_call' || event.type === 'tool_result') {
        store.setStatus('running')
      }
    })

    const offApproval = window.runtime.EventsOn('agent:approval_required', (data: unknown) => {
      store.setPendingApproval(data as ApprovalData)
      store.setStatus('waiting_approval')
    })

    return () => {
      offEvent()
      offApproval()
    }
  }, [])

  const loadThreads = useCallback(async () => {
    const threads = await wails().ListThreads()
    store.setThreads(threads ?? [])
  }, [])

  const newThread = useCallback(async () => {
    const id = await wails().NewThread('')
    await loadThreads()
    store.setActiveThread(id)
    return id
  }, [loadThreads])

  const deleteThread = useCallback(async (id: string) => {
    await wails().DeleteThread(id)
    store.removeThread(id)
    if (store.activeThreadId === id) {
      store.setActiveThread(null)
    }
  }, [store.activeThreadId])

  const sendMessage = useCallback(async (input: string) => {
    let threadId = store.activeThreadId
    if (!threadId) {
      threadId = await newThread()
    }

    store.clearChatEvents()
    store.setStatus('running')

    try {
      await wails().RunAgent(threadId, input)
    } catch (err) {
      store.setStatus('error')
      store.addChatEvent({
        id: `err-${Date.now()}`,
        event: {
          type: 'error',
          content: String(err),
          timestamp: Date.now(),
        },
      })
    }
  }, [store.activeThreadId, newThread])

  const stopAgent = useCallback(async () => {
    await wails().StopAgent()
    store.setStatus('idle')
  }, [])

  const approveAction = useCallback(async (approved: boolean) => {
    await wails().ApproveAction(approved)
    store.setPendingApproval(null)
    store.setStatus('running')
  }, [])

  const loadConfig = useCallback(async () => {
    const config = await wails().GetLLMConfig()
    store.setLLMConfig(config)
  }, [])

  const saveConfig = useCallback(async (config: import('../types').LLMConfig) => {
    await wails().SetLLMConfig(config)
    store.setLLMConfig(config)
  }, [])

  const getAvailableModels = useCallback(async () => {
    return wails().GetAvailableModels()
  }, [])

  return {
    loadThreads,
    newThread,
    deleteThread,
    sendMessage,
    stopAgent,
    approveAction,
    loadConfig,
    saveConfig,
    getAvailableModels,
  }
}
