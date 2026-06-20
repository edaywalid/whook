/** Hero visual: the core promise in one frame. A provider POSTs to whook,
 *  which returns a fast 200, saves the event durably, then delivers onward. */
export function CaptureFlow() {
  return (
    <svg
      viewBox="0 0 600 360"
      className="w-full"
      role="img"
      aria-label="Stripe sends a webhook to whook, which returns a fast 200 OK, writes the event to a durable store, and delivers it to your app."
    >
      <defs>
        <marker
          id="cf-arrow-slate"
          viewBox="0 0 10 10"
          refX="7"
          refY="5"
          markerWidth="5.5"
          markerHeight="5.5"
          orient="auto-start-reverse"
        >
          <path d="M0 0 10 5 0 10z" className="fill-http" />
        </marker>
        <marker
          id="cf-arrow-teal"
          viewBox="0 0 10 10"
          refX="7"
          refY="5"
          markerWidth="5.5"
          markerHeight="5.5"
          orient="auto-start-reverse"
        >
          <path d="M0 0 10 5 0 10z" className="fill-teal" />
        </marker>
      </defs>

      {/* return 200 arc (the fast ack) */}
      <path
        d="M268 120 C 232 64, 150 64, 116 142"
        fill="none"
        className="stroke-teal flow"
        strokeWidth="1.75"
        markerEnd="url(#cf-arrow-teal)"
      />
      <g className="font-mono">
        <rect x="160" y="44" width="64" height="24" rx="12" className="fill-teal-tint stroke-teal/30" strokeWidth="1" />
        <text x="192" y="60" textAnchor="middle" className="fill-teal-strong text-[12px] font-medium">
          200 OK
        </text>
      </g>

      {/* connectors */}
      <path d="M164 180 H231" fill="none" className="stroke-http flow" strokeWidth="1.75" markerEnd="url(#cf-arrow-slate)" />
      <path d="M369 180 H435" fill="none" className="stroke-http flow" strokeWidth="1.75" markerEnd="url(#cf-arrow-slate)" />
      <path d="M300 240 V258" fill="none" className="stroke-teal flow-slow" strokeWidth="1.75" markerEnd="url(#cf-arrow-teal)" />
      <text x="402" y="170" textAnchor="middle" className="fill-muted font-mono text-[11px]">
        deliver
      </text>

      {/* provider */}
      <g>
        <rect x="32" y="146" width="132" height="68" rx="14" className="fill-surface stroke-hairline" strokeWidth="1.5" />
        <text x="98" y="178" textAnchor="middle" className="fill-ink text-[15px] font-semibold">
          Stripe
        </text>
        <text x="98" y="197" textAnchor="middle" className="fill-muted font-mono text-[11px]">
          POST /webhook
        </text>
      </g>

      {/* whook */}
      <g>
        <rect x="231" y="120" width="138" height="120" rx="18" className="fill-teal" />
        <text x="300" y="158" textAnchor="middle" className="fill-white text-[18px] font-semibold">
          whook
        </text>
        <rect x="258" y="172" width="84" height="24" rx="12" className="breathe fill-white/15" />
        <text x="300" y="188" textAnchor="middle" className="fill-white font-mono text-[11px]">
          captured
        </text>
        <text x="300" y="220" textAnchor="middle" className="fill-teal-soft text-[12px]">
          saved to disk
        </text>
      </g>

      {/* durable store */}
      <g>
        <rect x="244" y="258" width="112" height="48" rx="12" className="fill-surface-2 stroke-hairline" strokeWidth="1.5" />
        <text x="300" y="280" textAnchor="middle" className="fill-ink text-[12px] font-medium">
          durable store
        </text>
        <g className="fill-teal">
          <circle cx="282" cy="294" r="2.5" />
          <circle cx="300" cy="294" r="2.5" />
          <circle cx="318" cy="294" r="2.5" />
        </g>
      </g>

      {/* your app */}
      <g>
        <rect x="436" y="146" width="132" height="68" rx="14" className="fill-surface stroke-hairline" strokeWidth="1.5" />
        <text x="502" y="178" textAnchor="middle" className="fill-ink text-[15px] font-semibold">
          Your app
        </text>
        <text x="502" y="197" textAnchor="middle" className="fill-muted font-mono text-[11px]">
          delivered
        </text>
      </g>
    </svg>
  )
}
