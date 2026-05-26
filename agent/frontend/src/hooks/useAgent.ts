import { useEffect, useRef } from 'react'
import { useAgentStore } from '../store'
import { ChatEvent } from '../types'

// Wails runtime - injected at runtime, declare to satisfy TypeScript
declare const window: Window & {
  runtime?: {
    EventsOn: (event: string, cb: (data: any) => void) => void
    EventsOff: (event: string) => void
  }
}

function eventsOn(event: string, cb: (data: any) => void) {
  // @ts-ignore - Wails injects this globally
  if (typeof window !== 'undefined' && (window as any).runtime) {
    // @ts-ignore
    ;(window as any).runtime.EventsOn(event, cb)
  } else {
    // fallback: try the wails runtime module path
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const { EventsOn } = require('../wailsjs/runtime/runtime')
      EventsOn(event, cb)
    } catch {
      // not in wails context
    }
  }
}

function eventsOff(event: string) {
  // @ts-ignore
  if (typeof window !== 'undefined' && (window as any).runtime) {
    // @ts-ignore
    ;(window as any).runtime.EventsOff(event)
  } else {
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const { EventsOff } = require('../wailsjs/runtime/runtime')
      EventsOff(event)
    } catch {
      // not in wails context
    }
  }
}

// Dynamically import Wails Go bindings to avoid TS module errors at compile time
async function getApp() {
  try {
    return await import('../wailsjs/go/app/App' as any)
  } catch {
    return null
  }
}

export function useAgent() {
  const addEvent = useAgentStore((s) => s.addEvent)
  const setThreads = useAgentStore((s) => s.setThreads)
  const setRunning = useAgentStore((s) => s.setRunning)
  const setPendingApproval = useAgentStore((s) => s.setPendingApproval)
  const activeThreadId = useAgentStore((s) => s.activeThreadId)
  const setActiveThread = useAgentStore((s) => s.setActiveThread)
  const setLLMConfig = useAgentStore((s) => s.setLLMConfig)

  const handlerRegistered = useRef(false)

  useEffect(() => {
    getApp().then((App) => {
      if (App) App.GetThreads().then(setThreads).catch(console.error)
    })

    if (handlerRegistered.current) return
    handlerRegistered.current = true

    eventsOn('agent:event', (payload: any) => {
      const { thread_id, type, content, data } = payload
      if (type === 'done') { setRunning(false); return }
      if (type === 'approval_required') { setPendingApproval(data); return }
      addEvent(thread_id, { type, content, data })
    })

    return () => {
      eventsOff('agent:event')
      handlerRegistered.current = false
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const newThread = async () => {
    const App = await getApp()
    if (!App) return
    const id = await App.NewThread('Nova conversa')
    const threads = await App.GetThreads()
    setThreads(threads)
    setActiveThread(id)
  }

  const deleteThread = async (id: string) => {
    const App = await getApp()
    if (!App) return
    await App.DeleteThread(id)
    const threads = await App.GetThreads()
    setThreads(threads)
    if (activeThreadId === id) setActiveThread(threads[0]?.id ?? null)
  }

  const sendMessage = async (message: string) => {
    if (!activeThreadId) return
    const App = await getApp()
    if (!App) return
    setRunning(true)
    addEvent(activeThreadId, { type: 'user', content: message, data: null })
    try {
      await App.SendMessage(activeThreadId, message)
    } catch (e) {
      addEvent(activeThreadId, { type: 'error', content: String(e), data: null })
      setRunning(false)
    }
  }

  const approveAction = async (approved: boolean) => {
    const App = await getApp()
    setPendingApproval(null)
    if (App) await App.ApproveAction(approved)
  }

  const saveConfig = async (config: any) => {
    const App = await getApp()
    if (App) await App.SaveConfig(config)
    setLLMConfig(config)
  }

  return { newThread, deleteThread, sendMessage, approveAction, saveConfig }
}
