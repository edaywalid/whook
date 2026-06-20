import { Reveal } from '../reveal'
import { ArchitectureFlow } from '../visuals/architecture-flow'

type Step = { n: string; title: string; body: string }

const STEPS: Step[] = [
  {
    n: '01',
    title: 'Capture',
    body: 'The raw request is written to disk the instant it arrives, then whook returns a fast 2xx.',
  },
  {
    n: '02',
    title: 'Store',
    body: 'Every event is kept durably with its headers, body, and source, queryable and replayable.',
  },
  {
    n: '03',
    title: 'Deliver',
    body: 'Workers forward to your services on their own schedule, retrying until success or dead-letter.',
  },
]

export function HowItWorks() {
  return (
    <section id="how" className="border-y border-hairline bg-surface-2">
      <div className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
        <Reveal>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-teal-bright">
            How it works
          </p>
          <h2 className="mt-4 max-w-[24ch] text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
            Capture is decoupled from delivery. That is the whole trick.
          </h2>
          <p className="mt-4 max-w-[58ch] text-lg leading-relaxed text-body">
            An inbound request is saved durably before it is acknowledged. Delivery
            happens afterward, on whook&rsquo;s own schedule, so a destination being
            down never costs you an event.
          </p>
        </Reveal>

        <Reveal delay={0.1} y={28}>
          <figure className="mt-12 rounded-2xl border border-hairline bg-surface px-4 py-8 shadow-[0_24px_50px_-30px_rgba(0,0,0,0.6)] sm:px-10">
            <ArchitectureFlow />
          </figure>
        </Reveal>

        <ol className="mt-12 grid gap-px overflow-hidden rounded-2xl border border-hairline bg-hairline md:grid-cols-3">
          {STEPS.map((step, i) => (
            <Reveal key={step.n} delay={i * 0.06}>
              <li className="flex h-full flex-col gap-3 bg-surface-2 p-7">
                <span className="font-mono text-sm text-teal">{step.n}</span>
                <h3 className="text-lg font-semibold text-ink">{step.title}</h3>
                <p className="text-[15px] leading-relaxed text-body">{step.body}</p>
              </li>
            </Reveal>
          ))}
        </ol>
      </div>
    </section>
  )
}
