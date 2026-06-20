import { createFileRoute } from '@tanstack/react-router'
import { DocArticle } from '../components/docs/doc-article'

export const Route = createFileRoute('/docs/$slug')({ component: DocsSlug })

function DocsSlug() {
  const { slug } = Route.useParams()
  return <DocArticle slug={slug} />
}
