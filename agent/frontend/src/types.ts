export type Thread = {
  id: string
  title: string
  created_at: string
  updated_at: string
}

export type LLMConfig = {
  provider: 'openai_compatible' | 'anthropic' | 'openai' | 'ollama'
  model: string
  api_key?: string
  base_url?: string
}

export type EventType =
  | 'user'
  | 'thinking'
  | 'message'
  | 'tool_call'
  | 'tool_result'
  | 'done'
  | 'error'
  | 'approval_required'

export type ToolCallData = {
  tool_name: string
  args: Record<string, unknown>
}

export type ToolResultData = {
  tool_name: string
  output: string
  success: boolean
}

export type AgentEventPayload = {
  type: EventType
  content: string
  data: ToolCallData | ToolResultData | null
}

export type AgentEvent = AgentEventPayload & {
  thread_id: string
}

export type ChatEvent = {
  id: string
  event: AgentEventPayload
}
