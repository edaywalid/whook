import {
  siDiscord,
  siGithub,
  siLinear,
  siPaypal,
  siShopify,
  siSquare,
  siStripe,
} from 'simple-icons'

const PROVIDERS = [
  siStripe,
  siGithub,
  siShopify,
  siSquare,
  siPaypal,
  siLinear,
  siDiscord,
]

export function ProviderStrip() {
  return (
    <section className="border-y border-hairline bg-surface-2">
      <div className="mx-auto max-w-[1200px] px-6 py-11">
        <p className="text-center font-mono text-xs uppercase tracking-[0.16em] text-muted">
          Built for the webhooks you already receive
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-x-12 gap-y-7">
          {PROVIDERS.map((icon) => (
            <svg
              key={icon.slug}
              role="img"
              aria-label={icon.title}
              viewBox="0 0 24 24"
              className="h-[26px] w-auto fill-muted opacity-70 transition-opacity duration-200 hover:opacity-100"
            >
              <path d={icon.path} />
            </svg>
          ))}
        </div>
      </div>
    </section>
  )
}
