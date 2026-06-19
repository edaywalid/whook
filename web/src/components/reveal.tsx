import { m, useReducedMotion } from 'motion/react'
import type { ReactNode } from 'react'

type RevealProps = {
  children: ReactNode
  delay?: number
  y?: number
  className?: string
}

/** Fade-and-rise on scroll into view. Uses the lazy `m` component (features are
 *  loaded once by LazyMotion in the root) and collapses to a plain render under
 *  reduced-motion so content is never gated behind an animation. */
export function Reveal({ children, delay = 0, y = 22, className }: RevealProps) {
  const reduce = useReducedMotion()
  return (
    <m.div
      className={className}
      initial={reduce ? false : { opacity: 0, y }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.25 }}
      transition={{ duration: 0.65, delay, ease: [0.16, 1, 0.3, 1] }}
    >
      {children}
    </m.div>
  )
}
