import { test, expect } from '@playwright/test';

// Assuming we have some login setup and admin access as in other tests

test.describe('Admin Problem Management', () => {
  // Use existing setup from other admin tests if possible, here using a direct flow assuming state is handled
  test('Create a new problem with all fields', async ({ page }) => {
    // 1. Login as admin
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'admin');
    await page.click('button[type="submit"]');
    
    // Ensure we are logged in
    await expect(page).toHaveURL('/');

    // 2. Go to Admin Create Problem
    await page.goto('/admin/problems/new');
    
    // Ensure there are no 500 errors
    await expect(page.locator('h1, h2, h3').filter({ hasText: 'Internal Server Error' })).toHaveCount(0);

    // 3. Fill the form
    await page.fill('input[name="code"]', 'e2e-test-problem');
    await page.fill('input[name="name"]', 'E2E Test Problem');
    await page.fill('input[name="category"]', 'E2E Testing');
    await page.fill('input[name="mirror_from"]', 'Playwright');
    
    await page.fill('textarea[name="description"]', 'This is a test description.');
    await page.fill('textarea[name="input_desc"]', 'Test Input');
    await page.fill('textarea[name="output_desc"]', 'Test Output');
    await page.fill('textarea[name="constraints_desc"]', 'Test Constraints');
    await page.fill('textarea[name="editorial_content"]', 'Test Editorial');

    await page.fill('input[name="time_limit"]', '2.0');
    await page.fill('input[name="memory_limit"]', '512');
    
    await page.selectOption('select[name="testcase_visibility"]', 'all');

    // Add tags
    await page.fill('#tagInput', 'e2e');
    await page.press('#tagInput', 'Enter');
    await page.fill('#tagInput', 'test');
    await page.press('#tagInput', 'Enter');
    
    // 4. Submit
    await page.click('button[type="submit"]');

    // 5. Verify creation (assuming it redirects to problems list or shows success)
    await expect(page.locator('h1, h2, h3').filter({ hasText: 'Internal Server Error' })).toHaveCount(0);
    // Ideally check if 'e2e-test-problem' is in the list
    await expect(page.locator('body')).toContainText('E2E Test Problem');
  });
});
