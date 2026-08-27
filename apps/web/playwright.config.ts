import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  retries: 1,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  expect: {
    toHaveScreenshot: {
      maxDiffPixelRatio: 0.05,
    },
  },
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'desktop',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 },
      },
    },
    {
      name: 'mobile',
      use: {
        ...devices['iPhone 13'],
        browserName: 'chromium',
      },
    },
  ],
  webServer: {
    command: 'E2E_TEST=true npx nuxi dev --port 3000',
    port: 3000,
    reuseExistingServer: true,
    timeout: 120_000,
  },
  outputDir: './e2e-results',
  snapshotDir: './e2e-snapshots',
})
