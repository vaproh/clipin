import { test, expect } from '@playwright/test'

test.describe('Landing page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1').first()).toBeVisible({ timeout: 15000 })
  })

  test('shows hero headline', async ({ page }) => {
    await expect(page.locator('h1').first()).toContainText('Clip. Post. Get Paid.')
  })

  test('shows eyebrow badge', async ({ page }) => {
    await expect(page.getByText('Performance Clipping Marketplace for India')).toBeVisible()
  })

  test('has navigation links on desktop', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'How it works' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'For Clippers' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'For Campaign Owners' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Why ClipIN' })).toBeVisible()
  })

  test('has CTA buttons', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'Get started' }).first()).toBeVisible()
  })

  test('marketplace preview shows campaigns', async ({ page }) => {
    await expect(page.getByText('Marketplace preview')).toBeVisible()
    await expect(page.getByText('Podcast XYZ')).toBeVisible()
    await expect(page.getByText('Business Creator')).toBeVisible()
    await expect(page.getByText('Gaming Podcast')).toBeVisible()
  })

  test('how it works section exists', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'How ClipIN Works' })).toBeVisible()
  })

  test('for clippers section exists', async ({ page }) => {
    await expect(page.getByText('Turn editing into a predictable side hustle')).toBeVisible()
  })

  test('for campaign owners section exists', async ({ page }) => {
    await expect(page.getByText('Turn one piece of content into distributed reach')).toBeVisible()
  })

  test('why clipin section exists', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Why ClipIN' })).toBeVisible()
  })

  test('faq section exists', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Frequently Asked Questions' })).toBeVisible()
  })

  test('final CTA section exists', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Ready to clip?' })).toBeVisible()
  })

  test('footer is visible', async ({ page }) => {
    await expect(page.locator('footer')).toBeVisible()
  })
})

test.describe('Landing page - Mobile', () => {
  test.use({ viewport: { width: 375, height: 812 } })

  test('hero renders on mobile', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1').first()).toContainText('Clip. Post. Get Paid.')
  })

  test('hamburger menu visible on mobile', async ({ page }) => {
    await page.goto('/')
    // The hamburger is a button with md:hidden - visible on mobile
    const buttons = page.locator('header button')
    await expect(buttons.first()).toBeVisible()
  })

  test('CTAs visible on mobile', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('link', { name: 'Get started' }).first()).toBeVisible()
  })
})

test.describe('Auth pages', () => {
  test('sign-in page loads', async ({ page }) => {
    await page.goto('/sign-in')
    // Page should load without SSR error
    await expect(page.locator('body')).toBeVisible()
    // Brand logo should be visible
    await expect(page.getByText('CI')).toBeVisible()
  })

  test('sign-up page loads', async ({ page }) => {
    await page.goto('/sign-up')
    await expect(page.locator('body')).toBeVisible()
    await expect(page.getByText('CI')).toBeVisible()
  })
})

test.describe('App shell redirects', () => {
  const appPages = ['/app', '/app/campaigns', '/app/submissions', '/app/earnings', '/app/payouts', '/app/settings']

  for (const path of appPages) {
    test(`${path} redirects to sign-in when unauthenticated`, async ({ page }) => {
      await page.goto(path)
      await page.waitForURL('**/sign-in**', { timeout: 10000 })
      expect(page.url()).toContain('/sign-in')
    })
  }
})
