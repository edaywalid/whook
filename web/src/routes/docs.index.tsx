import { createFileRoute } from '@tanstack/react-router'
import { DocArticle } from '../components/docs/doc-article'

export const Route = createFileRoute('/docs/')({ component: DocsIndex })

function DocsIndex() {
  return <DocArticle slug="introduction" />
}
