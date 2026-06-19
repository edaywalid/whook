import { Zap } from 'lucide-react'

export function Wordmark() {
  return (
    <span className="flex items-center gap-2">
      <span className="grid size-7 place-items-center rounded-md bg-teal text-white">
        <Zap className="size-4" strokeWidth={2.5} />
      </span>
      <span className="font-mono text-[17px] font-semibold tracking-tight text-ink">
        whook
      </span>
    </span>
  )
}
