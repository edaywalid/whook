import { createFileRoute } from '@tanstack/react-router'
import { SiteNav } from '../components/landing/site-nav'
import { Hero } from '../components/landing/hero'
import { ProviderStrip } from '../components/landing/provider-strip'
import { ProblemSection } from '../components/landing/problem-section'
import { HowItWorks } from '../components/landing/how-it-works'
import { Resilience } from '../components/landing/resilience'
import { FanOut } from '../components/landing/fan-out'
import { Features } from '../components/landing/features'
import { GettingStarted } from '../components/landing/getting-started'
import { Quickstart } from '../components/landing/quickstart'
import { DocsSection } from '../components/landing/docs-section'
import { CallToAction } from '../components/landing/call-to-action'
import { SiteFooter } from '../components/landing/site-footer'

export const Route = createFileRoute('/')({ component: Home })

function Home() {
  return (
    <div className="overflow-x-hidden">
      <SiteNav />
      <main>
        <Hero />
        <ProviderStrip />
        <ProblemSection />
        <HowItWorks />
        <Resilience />
        <FanOut />
        <Features />
        <GettingStarted />
        <Quickstart />
        <DocsSection />
        <CallToAction />
      </main>
      <SiteFooter />
    </div>
  )
}
