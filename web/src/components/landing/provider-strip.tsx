const PROVIDERS = [
  { name: 'Stripe', slug: 'stripe' },
  { name: 'GitHub', slug: 'github' },
  { name: 'Shopify', slug: 'shopify' },
  { name: 'Twilio', slug: 'twilio' },
  { name: 'Slack', slug: 'slack' },
  { name: 'Discord', slug: 'discord' },
] as const

export function ProviderStrip() {
  return (
    <section className="border-y border-hairline bg-surface-2">
      <div className="mx-auto max-w-[1200px] px-6 py-10">
        <p className="text-center font-mono text-xs uppercase tracking-[0.16em] text-muted">
          Built for the webhooks you already receive
        </p>
        <div className="mt-7 flex flex-wrap items-center justify-center gap-x-12 gap-y-7">
          {PROVIDERS.map((p) => (
            <img
              key={p.slug}
              src={`https://cdn.simpleicons.org/${p.slug}/6c7a76`}
              alt={p.name}
              width={28}
              height={28}
              loading="lazy"
              className="h-6 w-auto opacity-75 transition-opacity hover:opacity-100"
            />
          ))}
        </div>
      </div>
    </section>
  )
}
