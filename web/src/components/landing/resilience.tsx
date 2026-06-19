import { Database, RotateCcw, Zap } from 'lucide-react'
import { FeatureLine } from '../feature-line'
import { Reveal } from '../reveal'

export function Resilience() {
  return (
    <section className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <div className="grid gap-12 lg:grid-cols-2 lg:items-center lg:gap-16">
        <Reveal y={28} className="order-2 lg:order-1">
          <figure className="overflow-hidden rounded-2xl border border-hairline bg-surface p-4 shadow-[0_24px_70px_-40px_rgba(23,33,31,0.3)] sm:p-6">
            <img
              src="/diagrams/retry-timeline.png"
              alt="A failed delivery is retried with growing backoff gaps of one, two, four, and eight seconds, until an attempt returns 200 OK or the retry budget is exhausted and the event is dead-lettered."
              width={1600}
              height={900}
              className="w-full"
              loading="lazy"
            />
          </figure>
        </Reveal>
        <Reveal className="order-1 lg:order-2">
          <h2 className="text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
            Retries that back off, then dead-letter.
          </h2>
          <p className="mt-4 max-w-[52ch] text-lg leading-relaxed text-body">
            Transient failures are retried on a deterministic exponential schedule.
            A 429 honors its Retry-After. When the budget runs out, the delivery is
            dead-lettered so it stops burning resources and becomes visible.
          </p>
          <ul className="mt-7 space-y-3 text-[15px] text-body">
            <FeatureLine icon={Zap}>
              4xx that will never succeed are marked permanent, not retried.
            </FeatureLine>
            <FeatureLine icon={Database}>
              The pending queue lives in the database, so it survives a restart.
            </FeatureLine>
            <FeatureLine icon={RotateCcw}>
              Replay a dead-lettered event once you have fixed the cause.
            </FeatureLine>
          </ul>
        </Reveal>
      </div>
    </section>
  )
}
