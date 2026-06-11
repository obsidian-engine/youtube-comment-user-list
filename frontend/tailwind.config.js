/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        'c-bg': '#f7f5ef',
        'c-bg-2': '#ffffff',
        'c-ink': '#0a0a0f',
        'c-ink-dim': '#565664',
        'c-ink-mute': '#9a9aa6',
        'c-accent': '#e6004c',
        'c-accent-2': '#006c8a',
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
