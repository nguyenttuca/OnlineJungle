import { test, expect } from '@playwright/test';

test.describe('Console & Network Errors', () => {
  const urlsToTest = [
    '/',
    '/problems',
    '/problems?status=solved',
    '/problems?status=unsolved',
    '/problems?status=attempted',
    '/problems?q=a',
    '/contests',
    '/blogs',
    '/login'
  ];

  for (const url of urlsToTest) {
    test(`Check ${url} for console and network errors`, async ({ page }) => {
      const errors: string[] = [];
      const failedRequests: string[] = [];

      page.on('pageerror', error => errors.push(error.message));
      page.on('console', msg => {
        if (msg.type() === 'error') {
          errors.push(msg.text());
        }
      });
      
      page.on('response', response => {
        if (response.status() >= 500) {
          failedRequests.push(`${response.status()} ${response.url()}`);
        }
      });

      await page.goto(url);
      
      // Wait for network idle to catch delayed requests
      await page.waitForLoadState('networkidle');

      expect(errors, `Console errors found on ${url}`).toEqual([]);
      expect(failedRequests, `5xx Network errors found on ${url}`).toEqual([]);
    });
  }
});
