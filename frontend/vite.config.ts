import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: 'dist',
  },
  server: {
    port: 5173,
    proxy: {
      '/fields':       'http://localhost:8080',
      '/headlands':    'http://localhost:8080',
      '/swaths':       'http://localhost:8080',
      '/routes':       'http://localhost:8080',
      '/paths':        'http://localhost:8080',
      '/pipeline':     'http://localhost:8080',
      '/healthz':      'http://localhost:8080',
      '/readyz':       'http://localhost:8080',
      '/openapi.yaml': 'http://localhost:8080',
      '/docs':         'http://localhost:8080',
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
  },
})
