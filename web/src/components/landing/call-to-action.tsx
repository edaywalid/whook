import { ArrowUpRight, Github } from 'lucide-react'
import { REPO_URL } from '../../lib/constants'
import { Reveal } from '../reveal'

const CHIPS = ['Single Go binary', 'SQLite or Postgres', 'Prometheus metrics', 'MIT licensed']

export function CallToAction() {
  return (
    <section className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <Reveal y={28}>
        <div className="relative overflow-hidden rounded-3xl bg-surface-2 px-8 py-16 ring-1 ring-white/10 sm:px-16 sm:py-20">
          <div className="bg-grid pointer-events-none absolute inset-0 opacity-60" />
          <div className="relative grid items-center gap-12 lg:grid-cols-2">
            <div>
              <h2 className="max-w-[18ch] text-3xl font-semibold tracking-tight text-ink sm:text-5xl">
                Own your inbound webhooks.
              </h2>
              <p className="mt-5 max-w-[44ch] text-lg leading-relaxed text-body">
                Self-host whook next to your app. One binary, your database, your
                data. No accounts, no per-event pricing.
              </p>
              <div className="mt-8 flex flex-wrap items-center gap-3">
                <a
                  href={REPO_URL}
                  className="pressable inline-flex items-center gap-2 rounded-full bg-teal-bright px-6 py-3 text-sm font-semibold text-[#06201d] transition-colors hover:bg-[#5fe3d3]"
                >
                  <Github className="size-4" strokeWidth={2} />
                  Get whook on GitHub
                </a>
                <a
                  href="#quickstart"
                  className="pressable inline-flex items-center gap-2 rounded-full border border-white/15 px-6 py-3 text-sm font-medium text-ink transition-colors hover:border-white/30"
                >
                  Read the quickstart
                  <ArrowUpRight className="size-4" strokeWidth={2} />
                </a>
              </div>
            </div>

            <div className="lg:pl-8">
              <div className="rounded-2xl border border-white/10 bg-black/30 p-5">
                <div className="font-mono text-[13px] leading-relaxed">
                  <p className="text-muted"># pull, configure, run</p>
                  <p className="mt-1 text-ink/90">
                    <span className="text-teal-bright">$ </span>docker compose up -d
                  </p>
                  <p className="mt-1 text-[#5fb8a8]">whook listening on :8080</p>
                </div>
              </div>
              <ul className="mt-5 flex flex-wrap gap-2">
                {CHIPS.map((chip) => (
                  <li
                    key={chip}
                    className="rounded-full border border-white/10 px-3 py-1.5 font-mono text-[12px] text-body"
                  >
                    {chip}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </Reveal>
    </section>
  )
}
