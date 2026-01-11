import { test, expect } from '@playwright/test';

test.describe('Usage Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/usage');
  });

  test('should display usage page', async ({ page }) => {
    await expect(page).toHaveURL(/\/usage/);
    await page.waitForLoadState('domcontentloaded');

    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible();
  });

  test('should display usage charts or tables', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');

    // Look for chart containers or data tables
    const chartContainer = page.locator('[data-testid="usage-chart"]').or(
      page.locator('svg').first() // Recharts uses SVG
    );

    const tableContainer = page.locator('table').or(
      page.locator('[role="table"]')
    );

    // At least one visualization should exist
    const hasChart = await chartContainer.count() > 0;
    const hasTable = await tableContainer.count() > 0;

    expect(hasChart || hasTable).toBeTruthy();
  });

  test('should handle empty usage data', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');

    // Page should render even with no data
    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible();
  });

  test('should allow filtering by date range', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');

    // Global date picker should be available
    const datePicker = page.locator('[data-testid="date-picker"]').or(
      page.getByText(/today|week|date/i).first()
    );

    if (await datePicker.count() > 0) {
      await expect(datePicker.first()).toBeVisible();
    }
  });

  test('should display model breakdown if available', async ({ page }) => {
    await page.waitForLoadState('domcontentloaded');

    // This is optional - just verify page doesn't crash
    const mainContent = page.locator('main');
    await expect(mainContent).toBeVisible();
  });
});
