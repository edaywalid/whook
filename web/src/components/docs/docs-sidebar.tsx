import { Link } from '@tanstack/react-router'
import { DOCS_NAV } from '../../lib/docs'

const linkBase =
  'block rounded-md px-2.5 py-1.5 text-[14px] text-body transition-colors hover:text-ink'
const linkActive =
  'block rounded-md px-2.5 py-1.5 text-[14px] bg-teal-bright/10 text-teal-bright font-medium'

export function DocsSidebar() {
  return (
    <aside className="sticky top-16 hidden h-[calc(100dvh-4rem)] w-56 shrink-0 overflow-y-auto py-10 lg:block">
      <nav className="space-y-7">
        {DOCS_NAV.map((group) => (
          <div key={group.title}>
            <p className="mb-2 px-2.5 font-mono text-[11px] uppercase tracking-[0.14em] text-muted">
              {group.title}
            </p>
            <ul className="space-y-0.5">
              {group.items.map((item) =>
                item.slug === 'introduction' ? (
                  <li key={item.slug}>
                    <Link to="/docs" activeOptions={{ exact: true }} className={linkBase} activeProps={{ className: linkActive }}>
                      {item.title}
                    </Link>
                  </li>
                ) : (
                  <li key={item.slug}>
                    <Link to="/docs/$slug" params={{ slug: item.slug }} className={linkBase} activeProps={{ className: linkActive }}>
                      {item.title}
                    </Link>
                  </li>
                ),
              )}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  )
}
