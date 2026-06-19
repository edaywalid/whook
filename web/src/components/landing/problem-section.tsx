import { FileWarning, RotateCcw, ServerCrash, Split } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { Reveal } from '../reveal'

type Problem = { icon: LucideIcon; title: string; body: string }

const PROBLEMS: Problem[] = [
  {
    icon: ServerCrash,
    title: 'The event is gone when your app is down',
    body: 'A deploy, a crash, a slow restart. The provider sends once, retries weakly or not at all, and the "payment succeeded" never lands.',
  },
  {
    icon: FileWarning,
    title: 'Failures leave no trace',
    body: 'When processing throws, the raw request is already gone. You debug blind, with no record of what the provider actually sent.',
  },
  {
    icon: Split,
    title: 'One URL, many consumers',
    body: 'Billing, email, and analytics all want the same event, but the provider allows a single destination. So you hand-write fan-out glue.',
  },
  {
    icon: RotateCcw,
    title: 'No way to replay',
    body: 'You fixed the bug. Now you need yesterday’s event back, and there is no button for that.',
  },
]

export function ProblemSection() {
  return (
    <section className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
      <Reveal>
        <h2 className="max-w-[18ch] text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
          Pointing a provider straight at your app is fragile.
        </h2>
      </Reveal>
      <div className="mt-12 grid gap-px overflow-hidden rounded-2xl border border-hairline bg-hairline sm:grid-cols-2">
        {PROBLEMS.map((problem, i) => (
          <Reveal key={problem.title} delay={i * 0.05}>
            <article className="flex h-full flex-col gap-4 bg-surface p-7">
              <problem.icon className="size-6 text-dead" strokeWidth={1.75} />
              <h3 className="text-lg font-semibold text-ink">{problem.title}</h3>
              <p className="text-[15px] leading-relaxed text-body">
                {problem.body}
              </p>
            </article>
          </Reveal>
        ))}
      </div>
    </section>
  )
}
