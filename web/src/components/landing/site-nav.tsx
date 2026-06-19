import { Github } from 'lucide-react'
import { REPO_URL } from '../../lib/constants'
import { Wordmark } from '../wordmark'

const NAV_LINKS = [
  { label: 'How it works', href: '#how' },
  { label: 'Features', href: '#features' },
  { label: 'Quickstart', href: '#quickstart' },
] as const

export function SiteNav() {
  return (
    <header className="sticky top-0 z-40 border-b border-hairline bg-paper/85 backdrop-blur-md">
      <nav className="mx-auto flex h-16 max-w-[1200px] items-center justify-between px-6">
        <a href="#top" className="flex items-center gap-2">
          <Wordmark />
        </a>
        <div className="hidden items-center gap-8 md:flex">
          {NAV_LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm text-body transition-colors hover:text-ink"
            >
              {link.label}
            </a>
          ))}
        </div>
        <a
          href={REPO_URL}
          className="pressable inline-flex items-center gap-2 rounded-full border border-hairline bg-surface px-4 py-2 text-sm font-medium text-ink transition-colors hover:border-ink/25"
        >
          <Github className="size-4" strokeWidth={2} />
          GitHub
        </a>
      </nav>
    </header>
  )
}
