import type { ComponentType } from 'react'

export type DocLink = { slug: string; title: string }
export type DocGroup = { title: string; items: DocLink[] }

/** Sidebar structure for the docs site. `introduction` lives at /docs; the rest
 *  at /docs/<slug>. */
export const DOCS_NAV: DocGroup[] = [
  {
    title: 'Getting started',
    items: [
      { slug: 'introduction', title: 'Introduction' },
      { slug: 'quickstart', title: 'Quickstart' },
      { slug: 'installation', title: 'Installation' },
    ],
  },
  {
    title: 'Guides',
    items: [
      { slug: 'configuration', title: 'Configuration' },
      { slug: 'sources-and-destinations', title: 'Sources and destinations' },
      { slug: 'filters', title: 'Filters and fan-out' },
    ],
  },
  {
    title: 'Reference',
    items: [{ slug: 'http-api', title: 'HTTP API' }],
  },
  {
    title: 'Operations',
    items: [
      { slug: 'deployment', title: 'Deployment' },
      { slug: 'postgres', title: 'Postgres' },
    ],
  },
]

const DOCS_FLAT: DocLink[] = DOCS_NAV.flatMap((group) => group.items)

const modules = import.meta.glob('../content/docs/*.mdx', {
  eager: true,
}) as Record<string, { default: ComponentType }>

export function getDocComponent(slug: string): ComponentType | null {
  return modules[`../content/docs/${slug}.mdx`]?.default ?? null
}

/** Href for a doc: introduction is the docs root, others are /docs/<slug>. */
export function docHref(slug: string): string {
  return slug === 'introduction' ? '/docs' : `/docs/${slug}`
}

export function adjacentDocs(slug: string): { prev?: DocLink; next?: DocLink } {
  const i = DOCS_FLAT.findIndex((d) => d.slug === slug)
  if (i === -1) {
    return {}
  }
  return {
    prev: i > 0 ? DOCS_FLAT[i - 1] : undefined,
    next: i < DOCS_FLAT.length - 1 ? DOCS_FLAT[i + 1] : undefined,
  }
}
