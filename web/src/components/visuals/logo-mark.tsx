/** whook mark: an event dropping into a holding tray. The glyph uses
 *  currentColor so it inverts cleanly on light and dark tiles. */
export function LogoMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" className={className} aria-hidden="true">
      <path
        d="M4.5 12.5V15a4 4 0 0 0 4 4h7a4 4 0 0 0 4-4v-2.5"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <path
        d="M12 3.75v3"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <circle cx="12" cy="10.75" r="2.4" fill="currentColor" />
    </svg>
  )
}
