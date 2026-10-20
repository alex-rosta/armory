/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./internal/templates/**/*.html', './assets/js/**/*.js'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        alliance: '#0078FF',
        horde: '#B30000',
        // Gold accent
        primary: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
          950: '#451a03',
        },
      },
      backgroundImage: {
        'wow-bg': "url('/assets/media/wow-bg.png')",
      },
    },
  },
  plugins: [],
}
