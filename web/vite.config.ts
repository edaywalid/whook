import { defineConfig } from 'vite'
import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const config = defineConfig({
  plugins: [
    tailwindcss(),
    tanstackStart({
      // Prerender the landing page to static HTML so it can deploy as a static
      // site with no serverless runtime.
      prerender: { enabled: true, crawlLinks: true },
    }),
    viteReact(),
  ],
})

export default config
