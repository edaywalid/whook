/** The problem, drawn the same way as the solution diagrams: a provider sends
 *  a webhook straight at your app while it is deploying, the request has nowhere
 *  to land, and the event is lost. No retry, no record. */
export function WithoutWhook() {
  return (
    <svg
      viewBox="0 0 760 300"
      className="w-full"
      role="img"
      aria-label="Stripe sends a webhook straight at your app while it is deploying. The request has nowhere to land, the event falls away and is lost, and there is no retry and no record."
    >
      {/* incoming connector that never lands */}
      <path d="M188 132 H352" fill="none" className="stroke-http/70" strokeWidth="1.75" strokeDasharray="4 5" />
      <text x="270" y="120" textAnchor="middle" className="fill-muted font-mono text-[11px]">webhook</text>

      {/* break in the line */}
      <path d="M349 124 l8 16 M357 124 l8 16" className="stroke-dead" strokeWidth="2" strokeLinecap="round" />

      {/* fall trajectory of the lost event */}
      <path d="M360 134 C 366 178, 360 210, 384 230" fill="none" className="stroke-muted/50" strokeWidth="1.5" strokeDasharray="2 5" />

      {/* provider */}
      <rect x="36" y="100" width="150" height="64" rx="14" className="fill-surface stroke-hairline" strokeWidth="1.5" />
      <text x="111" y="130" textAnchor="middle" className="fill-ink text-[15px] font-semibold">Stripe</text>
      <text x="111" y="149" textAnchor="middle" className="fill-muted font-mono text-[11px]">POST /webhook</text>

      {/* your app, offline */}
      <rect x="574" y="100" width="150" height="64" rx="14" className="fill-surface-2 stroke-hairline" strokeWidth="1.5" strokeDasharray="5 5" />
      <circle cx="596" cy="124" r="4" className="fill-dead" />
      <text x="649" y="128" textAnchor="middle" className="fill-muted text-[15px] font-semibold">Your app</text>
      <text x="649" y="148" textAnchor="middle" className="fill-dead font-mono text-[11px]">deploying…</text>

      {/* the lost event */}
      <g transform="rotate(-8 388 250)">
        <rect x="324" y="230" width="128" height="40" rx="9" className="fill-paper stroke-dead/40" strokeWidth="1.5" />
        <text x="388" y="254" textAnchor="middle" className="fill-muted font-mono text-[12px]">payment.succeeded</text>
        <line x1="338" y1="250" x2="438" y2="250" className="stroke-dead/70" strokeWidth="1.5" />
      </g>
      <text x="470" y="262" className="fill-dead font-mono text-[12px] font-semibold tracking-[0.12em]">EVENT LOST</text>
      <text x="470" y="281" className="fill-muted font-mono text-[11px]">no retry · no record</text>
    </svg>
  )
}
