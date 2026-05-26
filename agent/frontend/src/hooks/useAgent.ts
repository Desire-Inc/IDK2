import { useEffect } from 'react'
import { useAgentStore } from '../store'

const wailsRuntime = () => (window as any).runtime as {
  EventsOn: (event: string, cb: (data: any) => void) => (() => void) | void
  EventsOff: (event: string) => void
} | undefined

const App = () => (window as any)?.go?.app?.App as {
  NewThread: (title: string) => Promise<string>
  GetThreads: () => Promise<any[]>
  DeleteThread: (id: string) => Promise<void>
  SendMessage: (threadID: string, message: string) => Promise<void>
  StopRun: (threadID: string) => Promise<void>
  ApproveAction: (approved: boolean) => Promise<void>
  SaveConfig: (cfg: any) => Promise<void>
  GetConfig: () => Promise<any>
} | undefined

let eventsRegistered = false
let initialLoadStarted = false

async function refreshThreads() {
  const threads = await App()?.GetThreads()
  if (threads) useAgentStore.getState().setThreads(threads)
}

function registerAgentEventsOnce() {
  if (eventsRegistered) return
  const runtime = wailsRuntime()
  if (!runtime) return

  eventsRegistered = true
  runtime.EventsOn('agent:event', (payload: any) => {
    const { thread_id, type, content, data } = payload
    const store = useAgentStore.getState()

    if (type === 'approval_required') {
      store.addEvent(thread_id, { type, content, data })
      store.setPendingApproval(data)
      return
    }

    if (type === 'title_updated') {
      refreshThreads().catch(console.error)
      return
    }

    store.addEvent(thread_id, { type, content, data })

    if (type === 'done' || type === 'cancelled' || type === 'error') {
      store.setRunning(false)
      refreshThreads().catch(console.error)
    }
  })
}

function loadInitialStateOnce() {
  if (initialLoadStarted) return
  initialLoadStarted = true
  App()?.GetThreads().then((threads) => {
    if (threads) useAgentStore.getState().setThreads(threads)
  }).catch(console.error)
  App()?.GetConfig().then((cfg) => {
    if (cfg) useAgentStore.getState().setLLMConfig(cfg)
  }).catch(console.error)
}

export function useAgent() {
  const activeId = useAgentStore((s) => s.activeThreadId)

  useEffect(() => {
    loadInitialStateOnce()
    registerAgentEventsOnce()
  }, [])

  const newThread = async () => {
    const app = App()
    if (!app) return
    const id = await app.NewThread('Nova conversa')
    const threads = await app.GetThreads()
    useAgentStore.getState().setThreads(threads)
    useAgentStore.getState().setActiveThread(id)
  }

  const deleteThread = async (id: string) => {
    const app = App()
    if (!app) return
    await app.DeleteThread(id)
    const threads = await app.GetThreads()
    const store = useAgentStore.getState()
    store.setThreads(threads)
    if (store.activeThreadId === id) store.setActiveThread(threads[0]?.id ?? null)
  }

  const sendMessage = async (message: string) => {
    const threadId = useAgentStore.getState().activeThreadId ?? activeId
    if (!threadId) return
    const app = App()
    if (!app) return

    const store = useAgentStore.getState()
    store.setRunning(true)
    store.addEvent(threadId, { type: 'user', content: message, data: null })

    try {
      await app.SendMessage(threadId, message)
    } catch (e) {
      store.addEvent(threadId, { type: 'error', content: String(e), data: null })
      store.setRunning(false)
    }
  }

  const stopRun = async () => {
    const threadId = useAgentStore.getState().activeThreadId ?? activeId
    if (!threadId) return
    await App()?.StopRun(threadId)
    useAgentStore.getState().setRunning(false)
  }

  const approveAction = async (approved: boolean) => {
    useAgentStore.getState().setPendingApproval(null)
    await App()?.ApproveAction(approved)
  }

  const saveConfig = async (config: any) => {
    await App()?.SaveConfig(config)
    useAgentStore.getState().setLLMConfig(config)
  }

  return { newThread, deleteThread, sendMessage, stopRun, approveAction, saveConfig }
}
