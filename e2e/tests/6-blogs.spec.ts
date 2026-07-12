import { test, expect } from '@playwright/test';

test.describe('Blogs', () => {
  test('View blogs list', async ({ page }) => {
    await page.goto('/blogs');
    await expect(page.locator('h1, h2, .blog-title').first()).toBeVisible();
  });
});
