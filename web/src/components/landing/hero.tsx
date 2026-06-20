import { ArrowRight, Github } from 'lucide-react'
import { REPO_URL } from '../../lib/constants'
import { Reveal } from '../reveal'
import { CaptureFlow } from '../visuals/capture-flow'

export function Hero() {
  return (
    <section id="top" className="relative overflow-hidden bg-grid">
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-px bg-gradient-to-r from-transparent via-hairline to-transparent" />
      <div className="relative mx-auto grid max-w-[1200px] gap-14 px-6 pt-20 pb-16 lg:grid-cols-12 lg:items-center lg:gap-10 lg:pt-24 lg:pb-24">
        <div className="lg:col-span-5">
          <Reveal>
            <span className="inline-flex items-center gap-2 rounded-full border border-teal-bright/25 bg-teal-bright/10 px-3 py-1 font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-teal-bright">
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
                className="pressable inline-flex items-center gap-2 rounded-full bg-teal-bright px-5 py-3 text-sm font-semibold text-[#06201d] transition-colors hover:bg-[#5fe3d3]"
              >
                Get started
                <ArrowRight className="size-4" strokeWidth={2} />
              </a>
              <a
                href={REPO_URL}
                className="pressable inline-flex items-center gap-2 rounded-full border border-white/12 bg-white/5 px-5 py-3 text-sm font-medium text-ink transition-colors hover:bg-white/10"
              >
                <Github className="size-4" strokeWidth={2} />
                View on GitHub
              </a>
            </div>
          </Reveal>
        </div>

        <div className="lg:col-span-7">
          <Reveal delay={0.2} y={28}>
            <figure className="rounded-2xl border border-hairline bg-surface p-5 shadow-[0_24px_50px_-30px_rgba(0,0,0,0.6)] sm:p-7">
              <figcaption className="mb-4 font-mono text-[11px] uppercase tracking-[0.14em] text-cmuted">
                capture &rarr; deliver
              </figcaption>
              <CaptureFlow />
            </figure>
          </Reveal>
        </div>
      </div>
    </section>
  )
}
