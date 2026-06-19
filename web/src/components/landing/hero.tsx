import { ArrowRight, Github } from 'lucide-react'
import { REPO_URL } from '../../lib/constants'
import { Reveal } from '../reveal'

export function Hero() {
  return (
    <section id="top" className="relative bg-grid">
      <div className="mx-auto grid max-w-[1200px] gap-14 px-6 pt-20 pb-16 lg:grid-cols-12 lg:items-center lg:gap-10 lg:pt-24 lg:pb-24">
        <div className="lg:col-span-5">
          <Reveal>
            <span className="inline-flex items-center gap-2 rounded-full border border-teal/25 bg-teal-tint px-3 py-1 font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-teal-strong">
              Self-hostable · Open source
            </span>
          </Reveal>
          <Reveal delay={0.06}>
            <h1 className="mt-6 text-[2.7rem] leading-[1.04] font-semibold tracking-[-0.03em] text-ink sm:text-6xl">
              Never lose a webhook again.
            </h1>
          </Reveal>
          <Reveal delay={0.12}>
            <p className="mt-5 max-w-[42ch] text-lg leading-relaxed text-body">
              whook captures every inbound webhook the instant it lands, returns a
              fast 2xx, then delivers with retries and replay.
            </p>
          </Reveal>
          <Reveal delay={0.18}>
            <div className="mt-8 flex flex-wrap items-center gap-3">
              <a
                href="#quickstart"
                className="pressable inline-flex items-center gap-2 rounded-full bg-teal px-5 py-3 text-sm font-medium text-white transition-colors hover:bg-teal-strong"
              >
                Get started
                <ArrowRight className="size-4" strokeWidth={2} />
              </a>
              <a
                href={REPO_URL}
                className="pressable inline-flex items-center gap-2 rounded-full border border-hairline bg-surface px-5 py-3 text-sm font-medium text-ink transition-colors hover:border-ink/25"
              >
                <Github className="size-4" strokeWidth={2} />
                View on GitHub
              </a>
            </div>
          </Reveal>
        </div>

        <div className="lg:col-span-7">
          <Reveal delay={0.2} y={28}>
            <figure className="overflow-hidden rounded-2xl border border-hairline bg-surface shadow-[0_24px_70px_-30px_rgba(23,33,31,0.4)]">
              <img
                src="/diagrams/without-vs-with-whook.png"
                alt="Without whook, a webhook sent while your app is deploying is lost. With whook, the request is captured to disk, acknowledged with a fast 2xx, and delivered to your app on schedule."
                width={1600}
                height={900}
                className="w-full"
                loading="eager"
              />
            </figure>
          </Reveal>
        </div>
      </div>
    </section>
  )
}
