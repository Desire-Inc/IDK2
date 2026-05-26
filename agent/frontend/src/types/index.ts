// ─── Agent Event Types ───────────────────────────────────────────────────────

export type EventType =
  | 'thinking'
  | 'tool_call'
  | 'tool_result'
  | 'message'
  | 'done'
  | 'error'
  | 'approval_required'

export interface AgentEvent {
  type: EventType
  content: string
  data?: ToolCallData | ToolResultData | ApprovalData
  timestamp: number
}

export interface ToolCallData {
  tool_name: string
  arguments: Record<string, unknown>
}

export interface ToolResultData {
  tool_name: string
  output: string
  error?: string
  success: boolean
}

export interface ApprovalData {
  approval_id: string
  action: string
  description: string
  data: unknown
}

// ─── Memory Types ─────────────────────────────────────────────────────────────

export interface Message {
  role: 'user' | 'assistant' | 'tool'
  content: string
  tool_id?: string
}

export interface Thread {
  id: string
  title: string
  created_at: string
  updated_at: string
}

// ─── LLM Config ───────────────────────────────────────────────────────────────

export interface LLMConfig {
  provider: 'anthropic' | 'openai' | 'ollama'
  model: string
  api_key?: string
  base_url?: string
}

// ─── UI State ─────────────────────────────────────────────────────────────────

export type AgentStatus = 'idle' | 'running' | 'waiting_approval' | 'error'

export interface ChatEvent {
  id: string
  event: AgentEvent
}
