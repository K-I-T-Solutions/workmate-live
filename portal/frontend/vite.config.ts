import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '0.0.0.0', // Listen on all network interfaces
    port: 5174,
    strictPort: false,
    hmr: {
      host: '192.168.178.100', // Your server IP for HMR
    },
    proxy: {
      '/api': {
        target: 'http://192.168.178.100:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://192.168.178.100:8080',
        ws: true,
        changeOrigin: true,
      },
    },
  },
})
