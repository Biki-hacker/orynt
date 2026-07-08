/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#f5f7ff',
          100: '#ebf0ff',
          500: '#3b82f6', // blue
          600: '#2563eb',
          700: '#1d4ed8',
          accent: '#4f46e5', // indigo
        },
        neutral: {
          50: '#fafafa', // off white
          100: '#f4f4f5', // light gray
          800: '#18181b', // near black
        }
      },
      fontFamily: {
        sans: ['Inter', 'Geist', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
