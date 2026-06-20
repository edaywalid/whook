/** Fan-out: one captured event resolves to independent delivery tracks, each
 *  with its own filter and status. One failing track never blocks the others. */
export function FanOutTracks() {
  return (
    <svg
      viewBox="0 0 600 360"
      className="w-full"
      role="img"
      aria-label="One captured event fans out to three independent tracks: billing is delivered, email is retrying, and analytics is dead-lettered. Each track has its own filter."
    >
      <defs>
        <marker id="fo-teal" viewBox="0 0 10 10" refX="7" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
          <path d="M0 0 10 5 0 10z" className="fill-teal" />
        </marker>
        <marker id="fo-warn" viewBox="0 0 10 10" refX="7" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
          <path d="M0 0 10 5 0 10z" className="fill-warn" />
        </marker>
        <marker id="fo-dead" viewBox="0 0 10 10" refX="7" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse">
          <path d="M0 0 10 5 0 10z" className="fill-dead" />
        </marker>
      </defs>

      {/* connectors */}
      <path d="M168 182 C 250 182, 250 70 332 70" fill="none" className="stroke-teal flow" strokeWidth="1.75" markerEnd="url(#fo-teal)" />
      <path d="M168 182 H332" fill="none" className="stroke-warn flow" strokeWidth="1.75" markerEnd="url(#fo-warn)" />
      <path d="M168 182 C 250 182, 250 294 332 294" fill="none" className="stroke-dead flow-slow" strokeWidth="1.5" strokeDasharray="2 5" markerEnd="url(#fo-dead)" />

      {/* event node */}
      <rect x="34" y="150" width="134" height="64" rx="14" className="fill-card stroke-cline" strokeWidth="1.5" />
      <text x="101" y="182" textAnchor="middle" className="fill-cink text-[15px] font-semibold">1 event</text>
      <text x="101" y="200" textAnchor="middle" className="fill-cmuted font-mono text-[11px]">order.created</text>

      <Track y={38} name="Billing" filter="amount > 0" status="delivered" tone="teal" />
      <Track y={150} name="Email" filter="type = receipt" status="retrying" tone="warn" />
      <Track y={262} name="Analytics" filter="*" status="dead-letter" tone="dead" />
    </svg>
  )
}

const TONES = {
  teal: { pill: 'fill-teal-tint', text: 'fill-teal-strong', dot: 'fill-teal' },
  warn: { pill: 'fill-warn/10', text: 'fill-warn', dot: 'fill-warn' },
  dead: { pill: 'fill-dead/10', text: 'fill-dead', dot: 'fill-dead' },
} as const

function Track({
  y,
  name,
  filter,
  status,
  tone,
}: {
  y: number
  name: string
  filter: string
  status: string
  tone: keyof typeof TONES
}) {
  const t = TONES[tone]
  return (
    <g>
      <rect x="332" y={y} width="234" height="64" rx="14" className="fill-card stroke-cline" strokeWidth="1.5" />
      <text x="352" y={y + 28} className="fill-cink text-[14px] font-semibold">
        {name}
      </text>
      <text x="352" y={y + 47} className="fill-cmuted font-mono text-[11px]">
        filter: {filter}
      </text>
      <g>
        <rect x={464} y={y + 14} width={92} height={22} rx={11} className={t.pill} />
        <circle cx={478} cy={y + 25} r={3} className={t.dot} />
        <text x={490} y={y + 29} className={`${t.text} font-mono text-[11px] font-medium`}>
          {status}
        </text>
      </g>
    </g>
  )
}
