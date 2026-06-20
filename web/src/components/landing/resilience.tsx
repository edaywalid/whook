import { Database, RotateCcw, Zap } from 'lucide-react'
import { FeatureLine } from '../feature-line'
import { Reveal } from '../reveal'
import { RetryTimeline } from '../visuals/retry-timeline'

export function Resilience() {
  return (
    <section className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <div className="grid gap-12 lg:grid-cols-2 lg:items-center lg:gap-16">
        <Reveal y={28} className="order-2 lg:order-1">
          <figure className="rounded-2xl bg-card px-4 py-8 shadow-[0_40px_90px_-30px_rgba(0,0,0,0.7),0_0_90px_-25px_rgba(45,212,191,0.3)] sm:px-7">
            <RetryTimeline />
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
