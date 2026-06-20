import type { ReactNode } from 'react'

type RevealProps = {
  children: ReactNode
  className?: string
  // Kept for call-site compatibility; the entrance is now CSS scroll-driven.
  delay?: number
  y?: number
}

/** Scroll-into-view entrance. Implemented with a CSS view() timeline (see
 *  styles.css) so content is visible by default and never gated behind JS. */
export function Reveal({ children, className }: RevealProps) {
  return <div className={`reveal ${className ?? ''}`}>{children}</div>
}
