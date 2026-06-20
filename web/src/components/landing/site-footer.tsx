import { REPO_URL } from '../../lib/constants'
import { Wordmark } from '../wordmark'

export function SiteFooter() {
  return (
    <footer className="border-t border-hairline">
      <div className="mx-auto flex max-w-[1200px] flex-col gap-6 px-6 py-12 sm:flex-row sm:items-center sm:justify-between">
        <Wordmark />
        <p className="text-sm text-muted">
          A self-hostable gateway for inbound webhooks. MIT licensed.
        </p>
        <div className="flex items-center gap-6 text-sm">
          <a href={REPO_URL} className="text-body transition-colors hover:text-ink">
            GitHub
          </a>
          <a href="/docs" className="text-body transition-colors hover:text-ink">
            Docs
          </a>
        </div>
      </div>
    </footer>
  )
}
