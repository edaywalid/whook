/** The full pipeline: provider to ingest (with a signature gate and the fast
 *  2xx ack) to a durable store, a router, delivery workers, and your services,
 *  plus the dead-letter and replay branches. */
export function ArchitectureFlow() {
  return (
    <svg
      viewBox="0 0 1000 360"
      className="w-full"
      role="img"
      aria-label="Provider to ingest, which verifies the signature and returns a fast 2xx, to a durable store, a router, delivery workers with retries, and your services. A dead-letter branch leaves the workers and a replay branch returns from the store."
    >
      <defs>
        <marker id="af-slate" viewBox="0 0 10 10" refX="7" refY="5" markerWidth="5.5" markerHeight="5.5" orient="auto-start-reverse">
          <path d="M0 0 10 5 0 10z" className="fill-http" />
        </marker>
        <marker id="af-teal" viewBox="0 0 10 10" refX="7" refY="5" markerWidth="5.5" markerHeight="5.5" orient="auto-start-reverse">
          <path d="M0 0 10 5 0 10z" className="fill-teal" />
        </marker>
        <marker id="af-dead" viewBox="0 0 10 10" refX="7" refY="5" markerWidth="5.5" markerHeight="5.5" orient="auto-start-reverse">
          <path d="M0 0 10 5 0 10z" className="fill-dead" />
        </marker>
      </defs>

      {/* main connectors */}
      <g strokeWidth="1.75" fill="none">
        <path d="M144 176 H180" className="stroke-http flow" markerEnd="url(#af-slate)" />
        <path d="M300 176 H336" className="stroke-http flow" markerEnd="url(#af-slate)" />
        <path d="M484 176 H520" className="stroke-teal flow" markerEnd="url(#af-teal)" />
        <path d="M640 176 H676" className="stroke-teal flow" markerEnd="url(#af-teal)" />
        <path d="M816 176 H852" className="stroke-teal flow" markerEnd="url(#af-teal)" />
      </g>

      {/* fast 2xx ack */}
      <path d="M240 140 C 222 90, 110 90, 84 140" fill="none" className="stroke-teal flow" strokeWidth="1.75" markerEnd="url(#af-teal)" />
      <rect x="120" y="66" width="84" height="24" rx="12" className="fill-teal-tint stroke-teal/30" strokeWidth="1" />
      <text x="162" y="82" textAnchor="middle" className="fill-teal-strong font-mono text-[11px] font-medium">fast 2xx</text>

      {/* replay branch (store -> workers) */}
      <path d="M410 222 C 430 300, 610 300, 700 224" fill="none" className="stroke-teal flow-slow" strokeWidth="1.5" strokeDasharray="2 5" markerEnd="url(#af-teal)" />
      <text x="556" y="296" textAnchor="middle" className="fill-teal-strong font-mono text-[11px]">replay</text>

      {/* dead-letter branch */}
      <path d="M746 222 V286" fill="none" className="stroke-dead" strokeWidth="1.5" strokeDasharray="2 5" markerEnd="url(#af-dead)" />

      {/* nodes */}
      <ArchNode x={24} label="Provider" sub="Stripe, GitHub…" />
      <g>
        <ArchNode x={180} label="Ingest" sub="capture + 2xx" />
        <rect x="186" y="116" width="108" height="20" rx="10" className="fill-surface stroke-hairline" strokeWidth="1" />
        <text x="240" y="130" textAnchor="middle" className="fill-muted font-mono text-[10px] uppercase tracking-wider">signature gate</text>
      </g>

      {/* durable store (focal) */}
      <g>
        <rect x="336" y="130" width="148" height="92" rx="16" className="fill-teal-tint stroke-teal/35" strokeWidth="1.5" />
        <text x="410" y="172" textAnchor="middle" className="fill-teal-strong text-[15px] font-semibold">Durable store</text>
        <text x="410" y="191" textAnchor="middle" className="fill-teal-strong/80 font-mono text-[11px]">saved before ack</text>
      </g>

      <ArchNode x={520} label="Router" sub="fan-out + filter" />

      {/* workers (taller) */}
      <g>
        <rect x="676" y="130" width="140" height="92" rx="16" className="fill-surface stroke-hairline" strokeWidth="1.5" />
        <text x="746" y="172" textAnchor="middle" className="fill-ink text-[15px] font-semibold">Delivery workers</text>
        <text x="746" y="191" textAnchor="middle" className="fill-muted font-mono text-[11px]">retry · backoff</text>
      </g>

      <ArchNode x={852} label="Your services" sub="billing, email…" w={124} />

      {/* dead-letter node */}
      <rect x="686" y="286" width="120" height="46" rx="12" className="fill-surface-2 stroke-dead/30" strokeWidth="1.5" />
      <text x="746" y="314" textAnchor="middle" className="fill-dead text-[12px] font-medium">dead-letter</text>
    </svg>
  )
}

function ArchNode({
  x,
  label,
  sub,
  w = 120,
}: {
  x: number
  label: string
  sub: string
  w?: number
}) {
  const cx = x + w / 2
  return (
    <g>
      <rect x={x} y={140} width={w} height={72} rx={14} className="fill-surface stroke-hairline" strokeWidth="1.5" />
      <text x={cx} y={172} textAnchor="middle" className="fill-ink text-[14px] font-semibold">
        {label}
      </text>
      <text x={cx} y={191} textAnchor="middle" className="fill-muted font-mono text-[10.5px]">
        {sub}
      </text>
    </g>
  )
}
