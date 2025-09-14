/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      gridTemplateColumns: {
        chips: "repeat(auto-fill,minmax(180px,1fr))",
      },
    },
  },
  plugins: [],
}
