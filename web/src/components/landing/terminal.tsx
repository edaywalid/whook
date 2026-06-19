type Tone = 'comment' | 'out' | 'cmd'
type Line = { id: string; text: string; tone: Tone }

const TERMINAL_LINES: Line[] = [
  { id: 'c1', tone: 'comment', text: '# 1. Register a source (your provider integration)' },
  { id: 'l1', tone: 'cmd', text: "curl -X POST localhost:8080/sources -d '{\"name\":\"stripe\"}'" },
  { id: 's1', tone: 'comment', text: '' },
  { id: 'c2', tone: 'comment', text: '# 2. Point it at one or more destinations' },
  { id: 'l2', tone: 'cmd', text: 'curl -X POST localhost:8080/destinations \\' },
  { id: 'l3', tone: 'cmd', text: "     -d '{\"source\":\"stripe\",\"url\":\"https://app.internal/hooks\"}'" },
  { id: 's2', tone: 'comment', text: '' },
  { id: 'c3', tone: 'comment', text: '# 3. Send a webhook to whook' },
  { id: 'l4', tone: 'cmd', text: 'curl -X POST localhost:8080/ingest/stripe \\' },
  { id: 'l5', tone: 'cmd', text: "     -d '{\"type\":\"payment.succeeded\",\"amount\":4900}'" },
  { id: 'o1', tone: 'out', text: '{"event_id":"evt_9f3a21c8b4e07d56"}' },
  { id: 's3', tone: 'comment', text: '' },
  { id: 'c4', tone: 'comment', text: '# 4. Inspect and replay anything that arrived' },
  { id: 'l6', tone: 'cmd', text: 'curl localhost:8080/events' },
]

export function Terminal() {
  return (
    <div className="overflow-hidden rounded-2xl border border-[#22302d] bg-[#0f1715] shadow-[0_30px_80px_-40px_rgba(15,23,21,0.7)]">
      <div className="flex items-center gap-2 border-b border-[#22302d] px-4 py-3">
        <span className="size-3 rounded-full bg-[#3a4744]" />
        <span className="size-3 rounded-full bg-[#3a4744]" />
        <span className="size-3 rounded-full bg-[#3a4744]" />
        <span className="ml-3 font-mono text-xs text-[#7c8a86]">bash</span>
      </div>
      <pre className="overflow-x-auto px-5 py-5 font-mono text-[13px] leading-[1.7]">
        <code>
          {TERMINAL_LINES.map((line) => (
            <span key={line.id} className="block">
              {line.tone === 'comment' ? (
                <span className="text-[#5f6f6b]">{line.text || ' '}</span>
              ) : line.tone === 'out' ? (
                <span className="tnum text-[#5fb8a8]">{line.text}</span>
              ) : (
                <span className="text-[#d7e2de]">
                  <span className="text-teal">$ </span>
                  {line.text}
                </span>
              )}
            </span>
          ))}
        </code>
      </pre>
    </div>
  )
}
