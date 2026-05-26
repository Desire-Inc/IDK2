import { useState, useRef, KeyboardEvent } from 'react'
import { Send, Square, ArrowUp } from 'lucide-react'
import clsx from 'clsx'

interface Props {
  onSubmit: (text: string) => void
  onStop: () => void
  disabled?: boolean
  running?: boolean
}

export default function Input({ onSubmit, onStop, disabled, running }: Props) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      submit()
    }
  }

  const submit = () => {
    const trimmed = value.trim()
    if (!trimmed || disabled) return
    onSubmit(trimmed)
    setValue('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleInput = () => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 200)}px`
  }

  return (
    <div
      className={clsx(
        'flex items-end gap-2 bg-notion-panel rounded-notion px-3 py-2 border border-notion-border',
        'focus-within:border-notion-accent/40 transition-colors',
        disabled && 'opacity-50'
      )}
    >
      <textarea
        ref={textareaRef}
        className="flex-1 bg-transparent resize-none text-sm text-notion-text placeholder-notion-muted outline-none min-h-[32px] max-h-[200px] py-0.5 leading-relaxed"
        placeholder="Digite uma tarefa para o agente..."
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onInput={handleInput}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        rows={1}
      />
      <button
        className={clsx(
          'flex-shrink-0 p-1.5 rounded-notion transition-colors mb-0.5',
          running
            ? 'bg-notion-red/20 hover:bg-notion-red/30 text-notion-red'
            : value.trim()
              ? 'bg-notion-accent text-white hover:bg-notion-accent/80'
              : 'text-notion-muted cursor-not-allowed'
        )}
        onClick={running ? onStop : submit}
        disabled={!running && !value.trim()}
      >
        {running ? <Square size={14} /> : <ArrowUp size={14} />}
      </button>
    </div>
  )
}
