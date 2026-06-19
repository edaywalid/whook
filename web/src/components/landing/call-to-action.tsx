import { Github } from 'lucide-react'
import { REPO_URL } from '../../lib/constants'
import { Reveal } from '../reveal'

export function CallToAction() {
  return (
    <section className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <Reveal y={28}>
        <div className="relative overflow-hidden rounded-3xl bg-ink px-8 py-16 text-center sm:px-16 sm:py-20">
          <div className="bg-grid pointer-events-none absolute inset-0 opacity-[0.06]" />
          <div className="relative">
            <h2 className="mx-auto max-w-[20ch] text-3xl font-semibold tracking-tight text-paper sm:text-5xl">
              Own your inbound webhooks.
            </h2>
            <p className="mx-auto mt-5 max-w-[48ch] text-lg leading-relaxed text-paper/70">
              Self-host it next to your app. One binary, your database, your data.
            </p>
            <div className="mt-9 flex flex-wrap items-center justify-center gap-3">
              <a
                href={REPO_URL}
                className="pressable inline-flex items-center gap-2 rounded-full bg-teal px-6 py-3 text-sm font-medium text-white transition-colors hover:bg-teal-strong"
              >
                <Github className="size-4" strokeWidth={2} />
                Get whook on GitHub
              </a>
              <a
                href="#quickstart"
                className="pressable inline-flex items-center gap-2 rounded-full border border-paper/20 px-6 py-3 text-sm font-medium text-paper transition-colors hover:border-paper/40"
              >
                Read the quickstart
              </a>
            </div>
          </div>
        </div>
      </Reveal>
    </section>
  )
}
