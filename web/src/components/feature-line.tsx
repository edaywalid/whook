import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'

export function FeatureLine({
  icon: Icon,
  children,
}: {
  icon: LucideIcon
  children: ReactNode
}) {
  return (
    <li className="flex items-start gap-3">
      <Icon className="mt-0.5 size-5 shrink-0 text-teal" strokeWidth={1.75} />
      <span>{children}</span>
    </li>
  )
}
