/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        'c-bg': '#f3f0e6',
        'c-bg-2': '#fdfcf8',
        'c-bg-3': '#ece7d8',
        'c-ink': '#0a0a0f',
        'c-ink-dim': '#4a4a56',
        'c-ink-mute': '#6a6a78',
        'c-accent': '#d6003f',
        'c-accent-2': '#005f78',
        'c-success': '#2d7a3f',
        'c-error': '#b3001b',
      },
      fontFamily: {
        display: ['"Zen Dots"', '"Zen Kaku Gothic New"', 'sans-serif'],
        body: ['"Zen Kaku Gothic New"', '"Hiragino Kaku Gothic ProN"', '"Yu Gothic"', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'ui-monospace', '"SF Mono"', 'Menlo', 'monospace'],
      },
      boxShadow: {
        subtle: '0 1px 2px 0 rgba(0,0,0,.05)',
      },
      gridTemplateColumns: {
        chips: 'repeat(auto-fill,minmax(180px,1fr))',
      },
    },
  },
  plugins: [],
}
