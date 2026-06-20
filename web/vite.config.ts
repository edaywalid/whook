import { defineConfig } from 'vite'
import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import mdx from '@mdx-js/rollup'
import remarkGfm from 'remark-gfm'
import rehypeSlug from 'rehype-slug'
import rehypeShiki from '@shikijs/rehype'

const config = defineConfig({
  plugins: [
    // MDX must run before the React plugin so .mdx becomes JSX first.
    {
      enforce: 'pre',
      ...mdx({
        remarkPlugins: [remarkGfm],
        rehypePlugins: [
          rehypeSlug,
          [rehypeShiki, { theme: 'github-dark-default' }],
        ],
      }),
    },
    tailwindcss(),
    tanstackStart({
      // Prerender the static site, crawling links so the landing page and every
      // docs page are emitted as static HTML.
      prerender: { enabled: true, crawlLinks: true },
    }),
    viteReact(),
  ],
})

export default config
