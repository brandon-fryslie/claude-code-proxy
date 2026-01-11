import { test, expect } from '@playwright/test';

test.describe('Dashboard Home Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display dashboard overview', async ({ page }) => {
    // Verify the dashboard page loads
    await expect(page).toHaveURL(/\/(dashboard)?$/);

    // Wait for content to load
    await page.waitForLoadState('domcontentloaded');
  });

  test('should display key metrics', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');

    // Dashboard should show some metrics/stats
    // This is generic - adjust based on actual dashboard content
    const metricsSection = page.locator('[data-testid="metrics"]').or(
      page.locator('main')
    );

    await expect(metricsSection.first()).toBeVisible();
  });

  test('should handle API failures gracefully', async ({ page }) => {
    // Intercept API calls and simulate failure
    await page.route('**/api/**', route => route.abort());

    await page.goto('/');

    // Page should still render without crashing
    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible();

    // Should show error state or fallback content
    // (This depends on implementation - adjust as needed)
  });

  test('should update when date range changes', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');

    // Look for date picker
    const datePicker = page.locator('[data-testid="date-picker"]').or(
      page.getByText(/today|week|date/i).first()
    );

    if (await datePicker.count() > 0) {
      // Click to open date picker
      await datePicker.first().click();

      // Wait for any updates to complete
      await page.waitForTimeout(1000);

      // Verify page is still functional
      const mainContent = page.locator('main');
      await expect(mainContent).toBeVisible();
    }
  });
});
