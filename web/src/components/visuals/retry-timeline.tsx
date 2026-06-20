/** Retry timeline: attempts fail with growing backoff gaps until one succeeds.
 *  A branch shows the alternative, a dead-letter once the budget is spent. */
export function RetryTimeline() {
  return (
    <svg
      viewBox="0 0 940 220"
      className="w-full"
      role="img"
      aria-label="Attempt one, two, and three fail with growing backoff gaps of two, six, and eighteen seconds. Attempt four returns 200 OK and is delivered. If the budget is exhausted instead, the delivery is dead-lettered."
    >
      {/* axis */}
      <line x1="70" y1="120" x2="870" y2="120" className="stroke-hairline" strokeWidth="2" />

      {/* backoff arcs */}
      <BackoffArc x1={110} x2={230} label="2s" />
      <BackoffArc x1={230} x2={400} label="6s" />
      <BackoffArc x1={400} x2={640} label="18s" />

      {/* dead-letter branch from attempt 3 */}
      <path d="M400 132 C 400 168, 380 176, 360 176" fill="none" className="stroke-dead" strokeWidth="1.5" strokeDasharray="2 5" />
      <rect x="240" y="176" width="120" height="34" rx="10" className="fill-surface-2 stroke-dead/30" strokeWidth="1.5" />
      <text x="300" y="197" textAnchor="middle" className="fill-dead text-[12px] font-medium">dead-letter</text>
      <text x="470" y="197" className="fill-muted font-mono text-[11px]">budget exhausted</text>

      <Attempt x={110} n="1" />
      <Attempt x={230} n="2" />
      <Attempt x={400} n="3" />
      <Success x={640} />
    </svg>
  )
}

function BackoffArc({ x1, x2, label }: { x1: number; x2: number; label: string }) {
  const mid = (x1 + x2) / 2
  return (
    <g>
      <path d={`M${x1} 120 Q ${mid} 80 ${x2} 120`} fill="none" className="stroke-http/50 flow-slow" strokeWidth="1.5" />
      <text x={mid} y={74} textAnchor="middle" className="fill-muted font-mono text-[11px]">
        {label}
      </text>
    </g>
  )
}

function Attempt({ x, n }: { x: number; n: string }) {
  return (
    <g>
      <circle cx={x} cy={120} r={12} className="fill-surface stroke-warn" strokeWidth="2" />
      <path
        d={`M${x - 4} 116 l8 8 M${x + 4} 116 l-8 8`}
        className="stroke-warn"
        strokeWidth="1.75"
        strokeLinecap="round"
      />
      <text x={x} y={158} textAnchor="middle" className="fill-muted font-mono text-[11px]">
        attempt {n}
      </text>
    </g>
  )
}

function Success({ x }: { x: number }) {
  return (
    <g>
      <circle cx={x} cy={120} r={14} className="fill-teal" />
      <path
        d={`M${x - 5.5} 120 l4 4 l7 -8`}
        fill="none"
        className="stroke-white"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <rect x={x + 28} y={108} width={150} height={26} rx={13} className="fill-teal-tint stroke-teal/30" strokeWidth="1" />
      <text x={x + 103} y={125} textAnchor="middle" className="fill-teal-strong font-mono text-[12px] font-medium">
        200 OK · delivered
      </text>
      <text x={x} y={158} textAnchor="middle" className="fill-teal-strong font-mono text-[11px]">
        attempt 4
      </text>
    </g>
  )
}
