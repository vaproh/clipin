import { test, expect } from '@playwright/test'

test.describe('Visual QA - Landing Desktop', () => {
  test('full page screenshot', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1').first()).toBeVisible({ timeout: 15000 })
    await expect(page).toHaveScreenshot('landing-desktop.png', {
      fullPage: true,
      maxDiffPixelRatio: 0.05,
    })
  })

  test('hero section', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1').first()).toBeVisible({ timeout: 15000 })
    await expect(page.locator('section').first()).toHaveScreenshot('hero-desktop.png', {
      maxDiffPixelRatio: 0.05,
    })
  })
})

test.describe('Visual QA - Landing Mobile', () => {
  test.use({ viewport: { width: 375, height: 812 } })

  test('full page screenshot', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('h1').first()).toBeVisible({ timeout: 15000 })
    await expect(page).toHaveScreenshot('landing-mobile.png', {
      fullPage: true,
      maxDiffPixelRatio: 0.05,
    })
  })
})

test.describe('Visual QA - Auth Pages', () => {
  test('sign-in page', async ({ page }) => {
    await page.goto('/sign-in')
    await expect(page).toHaveScreenshot('signin-desktop.png', {
      maxDiffPixelRatio: 0.05,
    })
  })

  test('sign-up page', async ({ page }) => {
    await page.goto('/sign-up')
    await expect(page).toHaveScreenshot('signup-desktop.png', {
      maxDiffPixelRatio: 0.05,
    })
  })
})

test.describe('Visual QA - Auth Mobile', () => {
  test.use({ viewport: { width: 375, height: 812 } })

  test('sign-in mobile', async ({ page }) => {
    await page.goto('/sign-in')
    await expect(page).toHaveScreenshot('signin-mobile.png', {
      maxDiffPixelRatio: 0.05,
    })
  })
})
