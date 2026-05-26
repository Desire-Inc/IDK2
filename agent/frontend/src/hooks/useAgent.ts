import { useEffect, useRef } from 'react'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import {
  NewThread,
  DeleteThread,
  SendMessage,
  ApproveAction,
  SaveConfig,
  GetThreads,
} from '../wailsjs/go/app/App'
import { useAgentStore } from '../store'
import { AgentEvent } from '../types'

export function useAgent() {
  const addEvent = useAgentStore((s) => s.addEvent)
  const setThreads = useAgentStore((s) => s.setThreads)
  const setRunning = useAgentStore((s) => s.setRunning)
  const setPendingApproval = useAgentStore((s) => s.setPendingApproval)
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const setActiveThread = useAgentStore((s) => s.setActiveThread)
  const setLLMConfig = useAgentStore((s) => s.setLLMConfig)

  // Track registered handler to avoid duplicates
  const handlerRegistered = useRef(false)

  useEffect(() => {
    // Load threads on mount
    GetThreads().then(setThreads).catch(console.error)

    // Register event handler only once
    if (handlerRegistered.current) return
    handlerRegistered.current = true

    EventsOn('agent:event', (payload: AgentEvent) => {
      const { thread_id, type, content, data } = payload

      if (type === 'done') {
        setRunning(false)
        return
      }

      if (type === 'approval_required') {
        setPendingApproval(data as any)
        return
      }

      addEvent(thread_id, { type, content, data })
    })

    return () => {
      EventsOff('agent:event')
      handlerRegistered.current = false
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const newThread = async () => {
    const id = await NewThread('Nova conversa')
    const threads = await GetThreads()
    setThreads(threads)
    setActiveThread(id)
  }

  const deleteThread = async (id: string) => {
    await DeleteThread(id)
    const threads = await GetThreads()
    setThreads(threads)
    if (activeThreadId === id) {
      setActiveThread(threads[0]?.id ?? null)
    }
  }

  const sendMessage = async (message: string) => {
    if (!activeThreadId) return
    setRunning(true)
    addEvent(activeThreadId, { type: 'user', content: message, data: null })
    try {
      await SendMessage(activeThreadId, message)
    } catch (e) {
      addEvent(activeThreadId, { type: 'error', content: String(e), data: null })
      setRunning(false)
    }
  }

  const approveAction = async (approved: boolean) => {
    setPendingApproval(null)
    await ApproveAction(approved)
  }

  const saveConfig = async (config: any) => {
    await SaveConfig(config)
    setLLMConfig(config)
  }

  return { newThread, deleteThread, sendMessage, approveAction, saveConfig }
}
