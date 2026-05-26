/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'SFMono-Regular', 'Consolas', 'monospace'],
      },
      colors: {
        notion: {
          bg:        '#191919',
          surface:   '#1e1e1e',
          panel:     '#252525',
          border:    'rgba(255,255,255,0.07)',
          text:      '#e6e6e5',
          muted:     '#9b9b9b',
          accent:    '#2383e2',
          green:     '#4daa57',
          red:       '#e03e3e',
          orange:    '#e9973f',
          hover:     'rgba(255,255,255,0.05)',
          selected:  'rgba(35,131,226,0.12)',
        },
        // Codex layout tokens mapped to the original Notion palette.
        // This keeps the agent-workspace layout without the bright "vibecoding" look.
        codex: {
          bg:        '#191919',
          sidebar:   '#191919',
          panel:     '#1e1e1e',
          card:      '#252525',
          card2:     '#1e1e1e',
          border:    'rgba(255,255,255,0.07)',
          border2:   'rgba(255,255,255,0.12)',
          text:      '#e6e6e5',
          muted:     '#9b9b9b',
          faint:     '#6f6f6f',
          accent:    '#2383e2',
          accent2:   '#2383e2',
          blue:      '#2383e2',
          green:     '#4daa57',
          red:       '#e03e3e',
          amber:     '#e9973f',
        },
      },
      boxShadow: {
        codex: '0 18px 48px rgba(0,0,0,0.28)',
        glow: '0 0 0 rgba(0,0,0,0)',
      },
      borderRadius: {
        notion: '8px',
        codex: '12px',
      },
      backgroundImage: {
        'codex-radial': 'linear-gradient(180deg, #191919 0%, #191919 100%)',
      },
    },
  },
  plugins: [],
}
