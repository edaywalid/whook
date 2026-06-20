import { Reveal } from '../reveal'

type Feature = { title: string; tag: string; body: string }

const FEATURES: Feature[] = [
  {
    title: 'Durable capture',
    tag: 'saved before 2xx',
    body: 'Every request is written to disk before the provider is acknowledged, so an event is never lost to a restart.',
  },
  {
    title: 'Backoff retries',
    tag: 'exponential + budget',
    body: 'Failed deliveries retry on a deterministic schedule until they succeed or exhaust the retry budget.',
  },
  {
    title: 'Dead-letter and replay',
    tag: 'recover anytime',
    body: 'Exhausted deliveries are set aside, queryable, and replayable once you have fixed the cause.',
  },
  {
    title: 'Signature verification',
    tag: 'Stripe, GitHub',
    body: 'Pluggable per-provider verifiers. Forged requests are captured for inspection but never delivered.',
  },
  {
    title: 'Fan-out with filters',
    tag: 'one event, many',
    body: 'An event resolves to many destinations, each with its own filter, retries, and dead-letter state.',
  },
  {
    title: 'Idempotent capture',
    tag: 'Idempotency-Key',
    body: 'A provider dedup key collapses re-sent events into one, even under a concurrent race.',
  },
  {
    title: 'Prometheus metrics',
    tag: 'GET /metrics',
    body: 'Ingest rate, write latency, delivery outcomes, and dead-letter volume, ready to scrape.',
  },
  {
    title: 'One static binary',
    tag: 'docker compose up',
    body: 'A single Go binary with SQLite built in. Postgres when you outgrow it. Nothing else to run.',
  },
]

export function Features() {
  return (
    <section id="features" className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <Reveal>
        <h2 className="max-w-[20ch] text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
          Everything you would otherwise hand-roll.
        </h2>
        <p className="mt-4 max-w-[54ch] text-lg leading-relaxed text-body">
          The pieces you normally bolt together yourself, in one process you run
          next to your app.
        </p>
      </Reveal>

      <div className="mt-14 grid gap-x-16 gap-y-12 sm:grid-cols-2">
        {FEATURES.map((feature) => (
          <Reveal key={feature.title}>
            <div className="flex items-baseline justify-between gap-4">
              <h3 className="text-lg font-semibold text-ink">{feature.title}</h3>
              <code className="shrink-0 font-mono text-[12px] text-teal-bright">
                {feature.tag}
              </code>
            </div>
            <p className="mt-3 text-[15px] leading-relaxed text-body">
              {feature.body}
            </p>
          </Reveal>
        ))}
      </div>
    </section>
  )
}
