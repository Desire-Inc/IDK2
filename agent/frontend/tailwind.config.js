/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      },
      colors: {
        // Notion dark palette
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
      },
      borderRadius: {
        notion: '8px',
      },
    },
  },
  plugins: [],
}
