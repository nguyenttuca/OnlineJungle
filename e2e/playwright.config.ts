import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 1, // specified by the user's prompt
  workers: 1, // run sequentially to avoid test DB collisions
  reporter: [['html', { open: 'never' }]],
  use: {
    baseURL: 'http://localhost:8085',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
