import { test, expect } from '@playwright/test';

test.describe('Submissions', () => {
  const user = 'testuser123';
  const pass = 'testuser123';

  test.beforeEach(async ({ page }) => {
    // Login as normal user
    await page.goto('/login');
    await page.fill('input[name="username"]', user);
    await page.fill('input[name="password"]', pass);
    await page.click('button[type="submit"]');
  });

  test('Submit solution and check history', async ({ page }) => {
    await page.goto('/problems/test-problem/submit');
    
    // Select language
    await page.selectOption('select[name="language_id"]', '54'); // 54 usually C++
    
    // Fill code (Monaco editor might be tricky, we can try to paste or set value via JS)
    // Or if there's a hidden textarea:
    // Let's just try to evaluate and set value if Monaco is used, or fallback to textarea
    const hasMonaco = await page.evaluate(() => typeof window.monaco !== 'undefined').catch(() => false);
    if (hasMonaco) {
      await page.evaluate(() => {
        window.editor.setValue('#include <iostream>\\nusing namespace std;\\nint main() { int a, b; cin >> a >> b; cout << a + b; return 0; }');
      });
    } else {
      // Find the hidden input or text area that stores code
      await page.evaluate(() => {
        const el = document.querySelector('input[name="code"], textarea[name="code"]');
        if (el) el.value = '#include <iostream>\\nusing namespace std;\\nint main() { int a, b; cin >> a >> b; cout << a + b; return 0; }';
      });
    }

    await page.click('button[type="submit"]');

    // Should redirect to submission detail or submissions list
    await expect(page).toHaveURL(/.*\/submissions.*/);
    
    // Since the Judge Node is an external service and not running locally,
    // the submission will remain in Pending/In Queue.
    // We just verify it appears in the list.
    await expect(page.locator('text=test-problem').first()).toBeVisible();
    await expect(page.locator('text=testuser123').first()).toBeVisible();
  });
});
