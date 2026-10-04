import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd())
  const target = `http://${env.VITE_API_HOST}:${env.VITE_API_PORT}`

  return {
    plugins: [react(), tailwindcss()],
    // Proxy API calls so the browser sees one origin: the refresh-token
    // cookie (SameSite=Strict, Path=/auth) is stored and sent without CORS.
    server: {
      proxy: {
        '/auth': target,
        '/pin': target,
        '/interaction': target,
      },
    },
  }
})
