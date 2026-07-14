import { test, expect } from '@playwright/test';

test.describe('Problem Search and Filter', () => {
  // Test search functionality
  test('Search for a specific problem', async ({ page }) => {
    await page.goto('/problems');
    
    // Fill the search input
    await page.fill('input[name="q"]', 'A+B');
    await page.press('input[name="q"]', 'Enter');

    // Wait for page to load and check if results are displayed without 500 error
    await expect(page.locator('h1, h2, h3').filter({ hasText: 'Internal Server Error' })).toHaveCount(0);
  });

  // Test status filters
  test('Filter problems by Solved status', async ({ page }) => {
    await page.goto('/problems?status=solved');
    
    // Check that there is no 500 error
    await expect(page.locator('h1, h2, h3').filter({ hasText: 'Internal Server Error' })).toHaveCount(0);
    
    // Check that the dropdown reflects the selected status
    await expect(page.locator('select[name="status"]')).toHaveValue('solved');
  });

  test('Filter problems by Unsolved status', async ({ page }) => {
    await page.goto('/problems?status=unsolved');
    
    // Check that there is no 500 error
    await expect(page.locator('h1, h2, h3').filter({ hasText: 'Internal Server Error' })).toHaveCount(0);
    
    // Check that the dropdown reflects the selected status
    await expect(page.locator('select[name="status"]')).toHaveValue('unsolved');
  });

  test('Filter problems by Attempted status', async ({ page }) => {
    await page.goto('/problems?status=attempted');
    
    // Check that there is no 500 error
    await expect(page.locator('h1, h2, h3').filter({ hasText: 'Internal Server Error' })).toHaveCount(0);
    
    // Check that the dropdown reflects the selected status
    await expect(page.locator('select[name="status"]')).toHaveValue('attempted');
  });
});
