import { Reveal } from '../reveal'

type Method = { label: string; note: string; cmd: string }

const METHODS: Method[] = [
  {
    label: 'Docker',
    note: 'prebuilt multi-arch image',
    cmd: 'docker run -p 8080:8080 -v whook-data:/data ghcr.io/edaywalid/whook:latest',
  },
  {
    label: 'Docker Compose',
    note: 'full stack from source',
    cmd: 'docker compose up --build -d',
  },
  {
    label: 'Go',
    note: 'requires Go 1.25+',
    cmd: 'go install github.com/edaywalid/whook/cmd/whook@latest',
  },
]

export function GettingStarted() {
  return (
    <section id="get-started" className="border-t border-hairline bg-surface-2">
      <div className="mx-auto max-w-[1200px] px-6 py-24 lg:py-28">
        <Reveal>
          <p className="font-mono text-xs uppercase tracking-[0.16em] text-teal-bright">
            Get started
          </p>
          <h2 className="mt-4 text-3xl font-semibold tracking-tight text-ink sm:text-4xl">
            Run it in under a minute.
          </h2>
          <p className="mt-4 max-w-[54ch] text-lg leading-relaxed text-body">
            Pick how you want to run whook. It listens on <code className="font-mono text-ink">:8080</code> and
            stores events in SQLite by default, no other services required.
          </p>
        </Reveal>

        <div className="mt-12 space-y-4">
          {METHODS.map((method) => (
            <Reveal key={method.label}>
              <div className="grid gap-4 rounded-2xl border border-hairline bg-surface p-5 sm:grid-cols-[180px_1fr] sm:items-center sm:p-6">
                <div>
                  <div className="text-[15px] font-semibold text-ink">{method.label}</div>
                  <div className="font-mono text-[12px] text-muted">{method.note}</div>
                </div>
                <code className="block overflow-x-auto rounded-lg border border-[#22302d] bg-[#0f1715] px-4 py-3 font-mono text-[13px] text-[#d7e2de]">
                  <span className="text-teal-bright">$ </span>
                  {method.cmd}
                </code>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  )
}
