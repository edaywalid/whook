import { ArrowUpRight } from 'lucide-react'
import { REPO_URL } from '../../lib/constants'
import { Reveal } from '../reveal'

type Doc = { title: string; href: string; desc: string }

const DOCS: Doc[] = [
  {
    title: 'Configuration',
    href: `${REPO_URL}/blob/main/docs/configuration.md`,
    desc: 'Every environment variable, plus auth, secrets, retention, and rate limiting.',
  },
  {
    title: 'HTTP API',
    href: `${REPO_URL}/blob/main/docs/http-api.md`,
    desc: 'Endpoints, sources, destinations, the filter spec, and replay.',
  },
  {
    title: 'Deployment',
    href: `${REPO_URL}/blob/main/docs/deployment.md`,
    desc: 'Docker, the GHCR image, Compose, and scaling out with Postgres.',
  },
  {
    title: 'Contributing',
    href: `${REPO_URL}/blob/main/CONTRIBUTING.md`,
    desc: 'How to build, test, and submit changes to whook.',
  },
]

export function DocsSection() {
  return (
    <section id="docs" className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <Reveal>
        <p className="font-mono text-xs uppercase tracking-[0.16em] text-teal-bright">
          Documentation
        </p>
        <h2 className="mt-4 text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
          Read the docs.
        </h2>
        <p className="mt-4 max-w-[54ch] text-lg leading-relaxed text-body">
          Reference for configuring, deploying, and integrating with whook.
        </p>
      </Reveal>

      <div className="mt-12 grid gap-x-16 gap-y-10 sm:grid-cols-2">
        {DOCS.map((doc) => (
          <Reveal key={doc.title}>
            <a
              href={doc.href}
              target="_blank"
              rel="noreferrer"
              className="group block border-t border-hairline pt-5 transition-colors hover:border-teal-bright/40"
            >
              <div className="flex items-center gap-1.5 text-lg font-semibold text-ink transition-colors group-hover:text-teal-bright">
                {doc.title}
                <ArrowUpRight className="size-4 text-muted transition-colors group-hover:text-teal-bright" strokeWidth={2} />
              </div>
              <p className="mt-2 text-[15px] leading-relaxed text-body">{doc.desc}</p>
            </a>
          </Reveal>
        ))}
      </div>
    </section>
  )
}
