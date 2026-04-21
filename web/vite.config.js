import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  base: './',
  plugins: [svelte()],
  server: {
    allowedHosts: true, // allow any host to access this page
    proxy: {
      '/api': {
        target: process.env.NODE_ENV === 'development' && process.env.DOCKER_ENV ? 
          'http://airstats:8080' : 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
