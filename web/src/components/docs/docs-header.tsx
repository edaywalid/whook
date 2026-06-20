import { Github } from 'lucide-react'
import { REPO_URL } from '../../lib/constants'
import { Wordmark } from '../wordmark'

export function DocsHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-hairline bg-paper/85 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-[1280px] items-center justify-between px-6">
        <a href="/" className="flex items-center gap-2.5">
          <Wordmark />
          <span className="font-mono text-[12px] uppercase tracking-[0.16em] text-muted">
            docs
          </span>
        </a>
        <a
          href={REPO_URL}
          className="pressable inline-flex items-center gap-2 rounded-full border border-white/12 bg-white/5 px-4 py-2 text-sm font-medium text-ink transition-colors hover:bg-white/10"
        >
          <Github className="size-4" strokeWidth={2} />
          GitHub
        </a>
      </div>
    </header>
  )
}
