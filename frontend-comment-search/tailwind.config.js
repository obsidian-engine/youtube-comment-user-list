/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        canvas: { light: '#f5f6f7', dark: '#0b0d10' },
      },
      boxShadow: {
        subtle: '0 1px 2px 0 rgba(0,0,0,.05)',
      },
    },
  },
  plugins: [],
}
