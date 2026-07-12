import { test, expect } from '@playwright/test';

test.describe('Problems', () => {
  test('List problems', async ({ page }) => {
    await page.goto('/problems');
    await expect(page.locator('text=A + B Problem')).toBeVisible();
  });

  test('Problem detail page', async ({ page }) => {
    await page.goto('/problems/test-problem');
    await expect(page.locator('text=Tính tổng 2 số nguyên $A$ và $B$')).toBeVisible();
    await expect(page.locator('text=Time Limit')).toBeVisible();
    await expect(page.locator('text=1000 ms')).toBeVisible();
  });
});
