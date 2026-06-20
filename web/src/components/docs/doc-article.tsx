import { ArrowLeft, ArrowRight } from 'lucide-react'
import { adjacentDocs, docHref, getDocComponent } from '../../lib/docs'

export function DocArticle({ slug }: { slug: string }) {
  const Content = getDocComponent(slug)
  if (!Content) {
    return (
      <div className="mx-auto max-w-[760px] py-10">
        <h1 className="text-2xl font-semibold text-ink">Page not found</h1>
        <p className="mt-3 text-body">
          That docs page does not exist. <a href="/docs" className="text-teal-bright">Back to the docs</a>.
        </p>
      </div>
    )
  }

  const { prev, next } = adjacentDocs(slug)

  return (
    <div className="mx-auto max-w-[760px]">
      <article className="prose docs-prose max-w-none">
        <Content />
      </article>

      <nav className="mt-16 flex items-stretch justify-between gap-4 border-t border-hairline pt-6">
        {prev ? (
          <a
            href={docHref(prev.slug)}
            className="group flex flex-col items-start rounded-lg border border-hairline px-4 py-3 transition-colors hover:border-teal-bright/40"
          >
            <span className="flex items-center gap-1 font-mono text-[11px] uppercase tracking-wider text-muted">
              <ArrowLeft className="size-3" /> Previous
            </span>
            <span className="mt-0.5 text-[15px] font-medium text-ink group-hover:text-teal-bright">
              {prev.title}
            </span>
          </a>
        ) : (
          <span />
        )}
        {next ? (
          <a
            href={docHref(next.slug)}
            className="group flex flex-col items-end rounded-lg border border-hairline px-4 py-3 text-right transition-colors hover:border-teal-bright/40"
          >
            <span className="flex items-center gap-1 font-mono text-[11px] uppercase tracking-wider text-muted">
              Next <ArrowRight className="size-3" />
            </span>
            <span className="mt-0.5 text-[15px] font-medium text-ink group-hover:text-teal-bright">
              {next.title}
            </span>
          </a>
        ) : (
          <span />
        )}
      </nav>
    </div>
  )
}
