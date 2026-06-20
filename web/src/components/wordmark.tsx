import { LogoMark } from './visuals/logo-mark'

export function Wordmark() {
  return (
    <span className="flex items-center gap-2.5">
      <span className="grid size-8 place-items-center rounded-[9px] bg-teal text-white shadow-[0_2px_8px_-2px_rgba(15,118,110,0.5)]">
        <LogoMark className="size-5" />
      </span>
      <span className="text-[18px] font-semibold tracking-[-0.02em] text-ink">
        whook
      </span>
    </span>
  )
}
