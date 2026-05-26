import { useEffect, useRef } from 'react'
import { useAgentStore } from '../store'

// Wails injects `window.runtime` at startup.
// We wrap calls so TypeScript is happy without needing node types.
const wailsRuntime = () => (window as any).runtime as {
  EventsOn: (event: string, cb: (data: any) => void) => void
  EventsOff: (event: string) => void
} | undefined

// Wails Go bindings are generated into frontend/wailsjs/go/app/App.js
// They are available globally as window["go"]["app"]["App"] at runtime.
const App = () => (window as any)?.go?.app?.App as {
  NewThread: (title: string) => Promise<string>
  GetThreads: () => Promise<any[]>
  DeleteThread: (id: string) => Promise<void>
  SendMessage: (threadID: string, message: string) => Promise<void>
  ApproveAction: (approved: boolean) => Promise<void>
  SaveConfig: (cfg: any) => Promise<void>
  GetConfig: () => Promise<any>
} | undefined

export function useAgent() {
  const addEvent      = useAgentStore((s) => s.addEvent)
  const setThreads    = useAgentStore((s) => s.setThreads)
  const setRunning    = useAgentStore((s) => s.setRunning)
  const setPending    = useAgentStore((s) => s.setPendingApproval)
  const activeId      = useAgentStore((s) => s.activeThreadId)
  const setActive     = useAgentStore((s) => s.setActiveThread)
  const setLLMConfig  = useAgentStore((s) => s.setLLMConfig)

  const registered = useRef(false)

  useEffect(() => {
    // Load threads on mount
    App()?.GetThreads().then(setThreads).catch(console.error)

    if (registered.current) return
    registered.current = true

    wailsRuntime()?.EventsOn('agent:event', (payload: any) => {
      const { thread_id, type, content, data } = payload
      if (type === 'done') { setRunning(false); return }
      if (type === 'approval_required') { setPending(data); return }
      addEvent(thread_id, { type, content, data })
    })

    return () => {
      wailsRuntime()?.EventsOff('agent:event')
      registered.current = false
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const newThread = async () => {
    const app = App()
    if (!app) return
    const id = await app.NewThread('Nova conversa')
    const threads = await app.GetThreads()
    setThreads(threads)
    setActive(id)
  }

  const deleteThread = async (id: string) => {
    const app = App()
    if (!app) return
    await app.DeleteThread(id)
    const threads = await app.GetThreads()
    setThreads(threads)
    if (activeId === id) setActive(threads[0]?.id ?? null)
  }

  const sendMessage = async (message: string) => {
    if (!activeId) return
    const app = App()
    if (!app) return
    setRunning(true)
    addEvent(activeId, { type: 'user', content: message, data: null })
    try {
      await app.SendMessage(activeId, message)
    } catch (e) {
      addEvent(activeId, { type: 'error', content: String(e), data: null })
      setRunning(false)
    }
  }

  const approveAction = async (approved: boolean) => {
    setPending(null)
    await App()?.ApproveAction(approved)
  }

  const saveConfig = async (config: any) => {
    await App()?.SaveConfig(config)
    setLLMConfig(config)
  }

  return { newThread, deleteThread, sendMessage, approveAction, saveConfig }
}
