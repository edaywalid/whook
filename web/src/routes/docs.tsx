import { Outlet, createFileRoute } from '@tanstack/react-router'
import { DocsHeader } from '../components/docs/docs-header'
import { DocsSidebar } from '../components/docs/docs-sidebar'

export const Route = createFileRoute('/docs')({ component: DocsLayout })

function DocsLayout() {
  return (
    <div className="min-h-[100dvh] bg-paper">
      <DocsHeader />
      <div className="mx-auto flex max-w-[1280px] gap-12 px-6">
        <DocsSidebar />
        <main className="min-w-0 flex-1 py-12 lg:py-16">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
