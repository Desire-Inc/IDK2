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
        codex: {
          bg:        '#0b0d10',
          sidebar:   '#101318',
          panel:     '#151922',
          card:      '#191f2a',
          card2:     '#11161d',
          border:    'rgba(148,163,184,0.16)',
          border2:   'rgba(148,163,184,0.24)',
          text:      '#eef2f8',
          muted:     '#8d98a8',
          faint:     '#5f6b7a',
          accent:    '#7c5cff',
          accent2:   '#21c7a8',
          blue:      '#5aa7ff',
          green:     '#4ade80',
          red:       '#fb7185',
          amber:     '#fbbf24',
        },
      },
      boxShadow: {
        codex: '0 24px 80px rgba(0,0,0,0.42)',
        glow: '0 0 48px rgba(124,92,255,0.18)',
      },
      borderRadius: {
        notion: '8px',
        codex: '16px',
      },
      backgroundImage: {
        'codex-radial': 'radial-gradient(circle at 50% -10%, rgba(124,92,255,0.20), transparent 34%), radial-gradient(circle at 85% 20%, rgba(33,199,168,0.10), transparent 26%)',
      },
    },
  },
  plugins: [],
}
