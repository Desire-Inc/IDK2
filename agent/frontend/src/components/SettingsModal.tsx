import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { X, Key, Cpu, Link } from 'lucide-react'
import { useAgentStore } from '../store'
import { useAgent } from '../hooks/useAgent'
import { LLMConfig } from '../types'

const PROVIDER_MODELS: Record<string, string[]> = {
  openai_compatible: ['mimo-v2.5-pro', 'gpt-4o', 'gpt-4o-mini'],
  anthropic: ['claude-sonnet-4-5', 'claude-opus-4-5', 'claude-haiku-4-5'],
  openai: ['gpt-4o', 'gpt-4o-mini', 'o3-mini'],
  ollama: ['llama3', 'mixtral', 'codestral'],
}

const PROVIDER_DEFAULT_URLS: Record<string, string> = {
  openai_compatible: 'https://opengateway.gitlawb.com/v1',
  ollama: 'http://localhost:11434',
}

export default function SettingsModal() {
  const setShowSettings = useAgentStore((s) => s.setShowSettings)
  const llmConfig = useAgentStore((s) => s.llmConfig)
  const { saveConfig } = useAgent()

  const [config, setConfig] = useState<LLMConfig>(llmConfig)

  useEffect(() => {
    setConfig(llmConfig)
  }, [llmConfig])

  const models = PROVIDER_MODELS[config.provider] ?? []
  const showBaseURL = config.provider === 'openai_compatible' || config.provider === 'ollama'
  const showAPIKey = config.provider !== 'ollama'

  const handleProviderChange = (provider: string) => {
    setConfig((c) => ({
      ...c,
      provider: provider as LLMConfig['provider'],
      model: PROVIDER_MODELS[provider]?.[0] ?? '',
      base_url: PROVIDER_DEFAULT_URLS[provider] ?? '',
    }))
  }

  const handleSave = async () => {
    await saveConfig(config)
    setShowSettings(false)
  }

  return (
    <AnimatePresence>
      <motion.div
        initial= opacity: 0 
        animate= opacity: 1 
        exit= opacity: 0 
        className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50"
        onClick={() => setShowSettings(false)}
      >
        <motion.div
          initial= scale: 0.95, opacity: 0 
          animate= scale: 1, opacity: 1 
          exit= scale: 0.95, opacity: 0 
          transition= duration: 0.15 
          className="bg-notion-surface rounded-notion shadow-2xl border border-notion-border w-[480px] p-6"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between mb-5">
            <h2 className="font-semibold text-notion-text">Configurações</h2>
            <button
              onClick={() => setShowSettings(false)}
              className="p-1 rounded hover:bg-notion-hover text-notion-muted transition-colors"
            >
              <X size={15} />
            </button>
          </div>

          {/* Provider */}
          <label className="block mb-4">
            <div className="flex items-center gap-1.5 text-xs text-notion-muted mb-1.5">
              <Cpu size={12} />
              Provedor LLM
            </div>
            <select
              className="w-full bg-notion-panel border border-notion-border rounded-notion px-3 py-2 text-sm text-notion-text outline-none focus:border-notion-accent transition-colors"
              value={config.provider}
              onChange={(e) => handleProviderChange(e.target.value)}
            >
              <option value="openai_compatible">Gitlawb Opengateway (padrão)</option>
              <option value="anthropic">Anthropic (Claude)</option>
              <option value="openai">OpenAI (GPT)</option>
              <option value="ollama">Ollama (Local)</option>
            </select>
          </label>

          {/* Model */}
          <label className="block mb-4">
            <div className="text-xs text-notion-muted mb-1.5">Modelo</div>
            <input
              className="w-full bg-notion-panel border border-notion-border rounded-notion px-3 py-2 text-sm text-notion-text outline-none focus:border-notion-accent transition-colors placeholder-notion-muted"
              value={config.model}
              onChange={(e) => setConfig((c) => ({ ...c, model: e.target.value }))}
              placeholder={PROVIDER_MODELS[config.provider]?.[0] ?? 'nome-do-modelo'}
              list="model-suggestions"
            />
            <datalist id="model-suggestions">
              {models.map((m) => <option key={m} value={m} />)}
            </datalist>
          </label>

          {/* Base URL (OpenAI-compatible / Ollama) */}
          {showBaseURL && (
            <label className="block mb-4">
              <div className="flex items-center gap-1.5 text-xs text-notion-muted mb-1.5">
                <Link size={12} />
                Endpoint URL
              </div>
              <input
                className="w-full bg-notion-panel border border-notion-border rounded-notion px-3 py-2 text-sm text-notion-text outline-none focus:border-notion-accent transition-colors placeholder-notion-muted"
                value={config.base_url ?? ''}
                onChange={(e) => setConfig((c) => ({ ...c, base_url: e.target.value }))}
                placeholder={PROVIDER_DEFAULT_URLS[config.provider] ?? 'https://...'}
              />
            </label>
          )}

          {/* API Key */}
          {showAPIKey && (
            <label className="block mb-5">
              <div className="flex items-center gap-1.5 text-xs text-notion-muted mb-1.5">
                <Key size={12} />
                API Key
              </div>
              <input
                type="password"
                className="w-full bg-notion-panel border border-notion-border rounded-notion px-3 py-2 text-sm text-notion-text outline-none focus:border-notion-accent transition-colors placeholder-notion-muted"
                placeholder="sua-api-key..."
                value={config.api_key ?? ''}
                onChange={(e) => setConfig((c) => ({ ...c, api_key: e.target.value }))}
              />
            </label>
          )}

          {/* Actions */}
          <div className="flex gap-2 justify-end">
            <button
              onClick={() => setShowSettings(false)}
              className="px-4 py-2 rounded-notion text-sm text-notion-muted hover:bg-notion-hover transition-colors"
            >
              Cancelar
            </button>
            <button
              onClick={handleSave}
              className="px-4 py-2 rounded-notion text-sm bg-notion-accent text-white hover:bg-notion-accent/80 transition-colors"
            >
              Salvar
            </button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  )
}
