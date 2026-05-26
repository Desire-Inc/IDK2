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

export type AgentStatus = 'idle' | 'running' | 'done' | 'error'
export type RiskLevel = 'safe' | 'medium' | 'dangerous'

export type ApprovalData = {
  action: string
  description: string
  risk?: RiskLevel
  args?: string
}

export type EventType =
  | 'user'
  | 'plan'
  | 'step'
  | 'thinking'
  | 'reflection'
  | 'message'
  | 'tool_call'
  | 'tool_result'
  | 'done'
  | 'cancelled'
  | 'error'
  | 'approval_required'
  | 'title_updated'

export type ToolCallData = {
  tool_name: string
  args: Record<string, unknown> | string
  arguments?: string
  risk?: RiskLevel
}

export type ToolResultData = {
  tool_name: string
  output: string
  success: boolean
  risk?: RiskLevel
}

export type GenericEventData = Record<string, unknown>

export type AgentEventPayload = {
  type: EventType
  content: string
  data: ToolCallData | ToolResultData | ApprovalData | GenericEventData | null
}

export type AgentEvent = AgentEventPayload & {
  thread_id: string
}

export type ChatEvent = {
  id: string
  event: AgentEventPayload
}
