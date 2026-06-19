import { Layers, Split } from 'lucide-react'
import { FeatureLine } from '../feature-line'
import { Reveal } from '../reveal'

export function FanOut() {
  return (
    <section className="border-t border-hairline bg-surface-2">
      <div className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
        <div className="grid gap-12 lg:grid-cols-2 lg:items-center lg:gap-16">
          <Reveal>
            <h2 className="text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
              One event, many destinations.
            </h2>
            <p className="mt-4 max-w-[52ch] text-lg leading-relaxed text-body">
              A captured event fans out to every matching destination as its own
              delivery track. Billing, email, and analytics each get their own
              status, so one failing consumer never blocks the others.
            </p>
            <ul className="mt-7 space-y-3 text-[15px] text-body">
              <FeatureLine icon={Split}>
                Per-destination filters route only the events each service wants.
              </FeatureLine>
              <FeatureLine icon={Layers}>
                Each track keeps its own attempt history and dead-letter state.
              </FeatureLine>
            </ul>
          </Reveal>
          <Reveal y={28}>
            <figure className="overflow-hidden rounded-2xl border border-hairline bg-surface p-4 shadow-[0_24px_70px_-40px_rgba(23,33,31,0.3)] sm:p-6">
              <img
                src="/diagrams/fan-out.png"
                alt="One captured event is delivered to three independent tracks: billing is delivered, email is retrying, and analytics is dead-lettered, each with its own filter tag and status."
                width={1600}
                height={900}
                className="w-full"
                loading="lazy"
              />
            </figure>
          </Reveal>
        </div>
      </div>
    </section>
  )
}
