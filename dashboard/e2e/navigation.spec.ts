import { test, expect } from '@playwright/test';

test.describe('Dashboard Navigation', () => {
  test('should load the dashboard home page', async ({ page }) => {
    await page.goto('/');

    // Wait for page to load
    await page.waitForLoadState('domcontentloaded');

    // Sidebar should be visible
    await expect(page.locator('nav')).toBeVisible();

    // Main content area should be visible
    await expect(page.locator('main')).toBeVisible();
  });

  test('should navigate between pages using sidebar', async ({ page }) => {
    await page.goto('/');

    // Click on Requests page
    await page.getByRole('button', { name: /requests/i }).click();
    await expect(page).toHaveURL(/\/requests/);

    // Click on Conversations page
    await page.getByRole('button', { name: /conversations/i }).click();
    await expect(page).toHaveURL(/\/conversations/);

    // Click on Token Usage page
    await page.getByRole('button', { name: /token usage/i }).click();
    await expect(page).toHaveURL(/\/usage/);

    // Click on Performance page
    await page.getByRole('button', { name: /performance/i }).click();
    await expect(page).toHaveURL(/\/performance/);

    // Click on Provider Routing page
    await page.getByRole('button', { name: /provider routing/i }).click();
    await expect(page).toHaveURL(/\/routing/);

    // Click on Settings page
    await page.getByRole('button', { name: /settings/i }).click();
    await expect(page).toHaveURL(/\/settings/);

    // Return to Dashboard
    await page.getByRole('button', { name: /dashboard/i }).click();
    await expect(page).toHaveURL(/\/(dashboard)?$/);
  });

  test('should have header with controls', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');

    // Header area should exist
    const header = page.locator('div').filter({ has: page.locator('button') }).first();
    await expect(header).toBeVisible();
  });

  test('should persist active sidebar item on page reload', async ({ page }) => {
    await page.goto('/');

    // Navigate to requests page
    await page.getByRole('button', { name: /requests/i }).click();
    await expect(page).toHaveURL(/\/requests/);

    // Reload the page
    await page.reload();

    // Should still be on requests page
    await expect(page).toHaveURL(/\/requests/);
  });
});
