/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        quiz: {
          bg:       '#0a0e1a',
          surface:  '#111827',
          border:   '#1f2937',
          accent:   '#3b82f6',
          gold:     '#f59e0b',
          red:      '#ef4444',
          green:    '#22c55e',
          text:     '#f3f4f6',
          muted:    '#6b7280',
        },
      },
      fontFamily: {
        display: ['"Bebas Neue"', 'Impact', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
