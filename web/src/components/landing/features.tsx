import { Activity, Boxes, Inbox, Network, ShieldCheck } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Reveal } from '../reveal'

type Feature = { icon: LucideIcon; title: string; body: string }

const FEATURES: Feature[] = [
  {
    icon: ShieldCheck,
    title: 'Signature verification',
    body: 'Pluggable per-provider verifiers for Stripe and GitHub. Forged requests are captured but never delivered.',
  },
  {
    icon: Network,
    title: 'Fan-out with filters',
    body: 'One event resolves to many destinations, each with its own filter, retry budget, and dead-letter state.',
  },
  {
    icon: Activity,
    title: 'Prometheus metrics',
    body: 'Ingest rate, write latency, delivery outcomes, and dead-letter volume, exposed at /metrics.',
  },
  {
    icon: Inbox,
    title: 'Idempotent capture',
    body: 'A dedup key from the provider collapses re-sent events into one, even under a concurrent race.',
  },
]

export function Features() {
  return (
    <section id="features" className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <Reveal>
        <h2 className="max-w-[20ch] text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
          Everything you would otherwise hand-roll.
        </h2>
      </Reveal>

      <div className="mt-12 grid gap-5 lg:grid-cols-3">
        <Reveal className="lg:col-span-1 lg:row-span-2">
          <article className="flex h-full flex-col justify-between gap-8 rounded-2xl bg-teal p-8 text-white">
            <Boxes className="size-7" strokeWidth={1.75} />
            <div>
              <h3 className="text-xl font-semibold">One static binary</h3>
              <p className="mt-3 text-[15px] leading-relaxed text-white/85">
                A single Go binary with SQLite built in. No runtime to install, no
                external services to operate. Postgres is there when you outgrow it.
              </p>
              <code className="mt-6 block rounded-lg bg-black/20 px-3 py-2 font-mono text-[13px] text-teal-soft">
                docker compose up -d
              </code>
            </div>
          </article>
        </Reveal>

        {FEATURES.map((feature, i) => (
          <Reveal key={feature.title} delay={i * 0.05}>
            <article className="flex h-full flex-col gap-3 rounded-2xl border border-hairline bg-surface p-7">
              <feature.icon className="size-6 text-teal" strokeWidth={1.75} />
              <h3 className="text-lg font-semibold text-ink">{feature.title}</h3>
              <p className="text-[15px] leading-relaxed text-body">{feature.body}</p>
            </article>
          </Reveal>
        ))}
      </div>
    </section>
  )
}
