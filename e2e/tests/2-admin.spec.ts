import { test, expect } from '@playwright/test';

test.describe('Admin actions', () => {
  const adminUser = 'toplearn';
  const adminPass = 'toplearn@admin';

  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('/login');
    await page.fill('input[name="username"]', adminUser);
    await page.fill('input[name="password"]', adminPass);
    await page.click('button[type="submit"]');
  });

  test('Create a regular user', async ({ page }) => {
    await page.goto('/admin/users/create');
    await page.fill('input[name="username"]', 'testuser123');
    await page.fill('input[name="password"]', 'testuser123');
    await page.fill('input[name="display_name"]', 'Test User');
    await page.selectOption('select[name="role"]', 'user');
    await page.click('button[type="submit"]');

    await expect(page.locator('.alert-success')).toContainText('User created successfully');
  });

  test('Create a problem', async ({ page }) => {
    await page.goto('/admin/problems/create');
    await page.fill('input[name="slug"]', 'test-problem');
    await page.fill('input[name="title"]', 'A + B Problem');
    await page.fill('textarea[name="statement"]', 'Tính tổng 2 số nguyên $A$ và $B$.');
    await page.fill('input[name="time_limit"]', '1000'); // ms
    await page.fill('input[name="memory_limit"]', '256'); // mb
    
    // Create
    await page.click('button[type="submit"]');

    await expect(page.locator('.alert-success')).toBeVisible();

    // Upload test cases
    await page.goto('/admin/problems/test-problem/tests');
    await page.setInputFiles('input[type="file"]', 'fixtures/tests.zip');
    await page.click('button[type="submit"]');
    await expect(page.locator('.alert-success')).toBeVisible();
  });

  test('Create a contest', async ({ page }) => {
    await page.goto('/admin/contests/create');
    await page.fill('input[name="title"]', 'Test Contest');
    
    // Simple dates
    const start = new Date();
    start.setMinutes(start.getMinutes() - 10); // started 10 mins ago
    
    const end = new Date();
    end.setHours(end.getHours() + 2); // ends in 2 hours

    // Format for input type="datetime-local": YYYY-MM-DDThh:mm
    const formatDt = (d: Date) => {
      return d.toISOString().slice(0,16);
    };

    await page.fill('input[name="start_at"]', formatDt(start));
    await page.fill('input[name="end_at"]', formatDt(end));
    // Set ranking type
    await page.selectOption('select[name="ranking_type"]', 'ICPC');

    await page.click('button[type="submit"]');
    
    await expect(page.locator('.alert-success')).toBeVisible();
  });
});
