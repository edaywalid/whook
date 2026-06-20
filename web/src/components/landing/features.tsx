import { Activity, Boxes, Inbox, Network, ShieldCheck } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Reveal } from '../reveal'

type Feature = { icon: LucideIcon; title: string; body: string; accent: string }

const FEATURES: Feature[] = [
  {
    icon: ShieldCheck,
    title: 'Signature verification',
    body: 'Pluggable per-provider verifiers for Stripe and GitHub. Forged requests are captured but never delivered.',
    accent: 'X-Hub-Signature-256',
  },
  {
    icon: Network,
    title: 'Fan-out with filters',
    body: 'One event resolves to many destinations, each with its own filter, retry budget, and dead-letter state.',
    accent: 'type = receipt',
  },
  {
    icon: Activity,
    title: 'Prometheus metrics',
    body: 'Ingest rate, write latency, delivery outcomes, and dead-letter volume, exposed for scraping.',
    accent: 'GET /metrics',
  },
  {
    icon: Inbox,
    title: 'Idempotent capture',
    body: 'A dedup key from the provider collapses re-sent events into one, even under a concurrent race.',
    accent: 'Idempotency-Key',
  },
]

const INSIDE = ['ingest', 'store', 'router', 'workers', 'metrics']

export function Features() {
  return (
    <section id="features" className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <Reveal>
        <h2 className="max-w-[20ch] text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
          Everything you would otherwise hand-roll.
        </h2>
        <p className="mt-4 max-w-[52ch] text-lg leading-relaxed text-body">
          The pieces you would normally bolt together yourself, in one process you
          run next to your app.
        </p>
      </Reveal>

      <div className="mt-12 grid gap-5 lg:grid-cols-3">
        <Reveal className="lg:row-span-2">
          <article className="flex h-full flex-col gap-6 rounded-2xl bg-teal p-8 text-white">
            <span className="grid size-11 place-items-center rounded-xl bg-white/15">
              <Boxes className="size-6" strokeWidth={1.75} />
            </span>
            <div>
              <h3 className="text-xl font-semibold text-white">One static binary</h3>
              <p className="mt-3 text-[15px] leading-relaxed text-white/85">
                A single Go binary with SQLite built in. No runtime to install, no
                external services to operate. Postgres is there when you outgrow it.
              </p>
            </div>
            <div className="mt-auto">
              <p className="font-mono text-[11px] uppercase tracking-[0.16em] text-teal-soft/70">
                everything inside
              </p>
              <div className="mt-2.5 flex flex-wrap gap-1.5">
                {INSIDE.map((part) => (
                  <span
                    key={part}
                    className="rounded-md bg-white/10 px-2 py-1 font-mono text-[12px] text-white/90"
                  >
                    {part}
                  </span>
                ))}
              </div>
              <code className="mt-4 block rounded-lg bg-black/25 px-3 py-2.5 font-mono text-[13px] text-teal-soft">
                <span className="text-white/40">$ </span>docker compose up -d
              </code>
            </div>
          </article>
        </Reveal>

        {FEATURES.map((feature, i) => (
          <Reveal key={feature.title} delay={i * 0.05}>
            <article className="flex h-full flex-col gap-4 rounded-2xl border border-hairline bg-surface p-7 transition-colors hover:border-teal/40">
              <span className="grid size-10 place-items-center rounded-xl bg-teal-tint text-teal">
                <feature.icon className="size-5" strokeWidth={1.75} />
              </span>
              <div>
                <h3 className="text-lg font-semibold text-ink">{feature.title}</h3>
                <p className="mt-2 text-[15px] leading-relaxed text-body">
                  {feature.body}
                </p>
              </div>
              <code className="mt-auto w-fit rounded-md bg-surface-2 px-2 py-1 font-mono text-[12px] text-muted">
                {feature.accent}
              </code>
            </article>
          </Reveal>
        ))}
      </div>
    </section>
  )
}
