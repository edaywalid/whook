import { ArrowRight } from 'lucide-react'
import { Reveal } from '../reveal'
import { Terminal } from './terminal'

export function Quickstart() {
  return (
    <section id="quickstart" className="border-t border-hairline bg-surface-2">
      <div className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
        <div className="grid gap-12 lg:grid-cols-12 lg:gap-16">
          <Reveal className="lg:col-span-4">
            <p className="font-mono text-xs uppercase tracking-[0.16em] text-teal-bright">
              Quickstart
            </p>
            <h2 className="mt-4 text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
              Point your provider at whook instead.
            </h2>
            <p className="mt-4 text-lg leading-relaxed text-body">
              Register a source, add a destination, and send. Everything that comes
              in is captured, listed, and replayable.
            </p>
            <a
              href="/docs"
              className="pressable mt-7 inline-flex items-center gap-2 text-sm font-medium text-teal-bright hover:text-teal"
            >
              Read the docs
              <ArrowRight className="size-4" strokeWidth={2} />
            </a>
          </Reveal>

          <Reveal delay={0.1} y={24} className="lg:col-span-8">
            <Terminal />
          </Reveal>
        </div>
      </div>
    </section>
  )
}
