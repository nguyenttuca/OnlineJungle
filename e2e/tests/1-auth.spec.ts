import { test, expect } from '@playwright/test';

test.describe('Auth', () => {
  // Use the seeded admin account
  const adminUser = 'toplearn';
  const adminPass = 'toplearn@admin';

  test('Login successful', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', adminUser);
    await page.fill('input[name="password"]', adminPass);
    await page.click('button[type="submit"]');

    // Should redirect to home or somewhere logged in
    await expect(page).toHaveURL('/');
    // Check if the user's name is visible in navbar
    await expect(page.locator('.dropdown-toggle').filter({ hasText: adminUser })).toBeVisible();
  });

  test('Login with wrong password', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', adminUser);
    await page.fill('input[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    // Expect an error message
    await expect(page.locator('text=Invalid credentials')).toBeVisible();
  });

  test('Login with non-existent user', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'nonexistent_user_123');
    await page.fill('input[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    await expect(page.locator('text=Invalid credentials')).toBeVisible();
  });

  test('Logout', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[name="username"]', adminUser);
    await page.fill('input[name="password"]', adminPass);
    await page.click('button[type="submit"]');
    
    // Logout by visiting /logout directly to bypass dropdown UI interaction
    await page.goto('/logout');
    await expect(page.locator('text=Login').first()).toBeVisible();
  });

  test('Redirect when accessing protected route', async ({ page }) => {
    // Try to access admin page without logging in
    await page.goto('/admin');
    await expect(page).toHaveURL(/.*\/login.*/);
  });
});
