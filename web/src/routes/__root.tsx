import {
  HeadContent,
  Outlet,
  Scripts,
  createRootRoute,
} from '@tanstack/react-router'
import { LazyMotion, domAnimation } from 'motion/react'
import type { ReactNode } from 'react'

import '@fontsource-variable/geist'
import '@fontsource-variable/jetbrains-mono'
import '../styles.css'

export const Route = createRootRoute({
  head: () => ({
    meta: [
      { charSet: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { title: 'whook: durable webhook gateway' },
      {
        name: 'description',
        content:
          'whook captures every inbound webhook the instant it lands, returns a fast 2xx, then delivers to your services with retries, dead-letter, and replay. One self-hostable Go binary.',
      },
      { name: 'theme-color', content: '#0f766e' },
      { property: 'og:title', content: 'whook: durable webhook gateway' },
      {
        property: 'og:description',
        content:
          'Capture, retry, and replay inbound webhooks. One self-hostable Go binary.',
      },
      { property: 'og:type', content: 'website' },
    ],
    links: [
      { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
    ],
  }),
  component: RootComponent,
})

function RootComponent() {
  return (
    <RootDocument>
      <LazyMotion features={domAnimation} strict>
        <Outlet />
      </LazyMotion>
    </RootDocument>
  )
}

function RootDocument({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  )
}
