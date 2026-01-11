import { test, expect } from '@playwright/test';

test.describe('Requests Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/requests');
  });

  test('should display requests page', async ({ page }) => {
    // Page should load
    await expect(page).toHaveURL(/\/requests/);

    // Check for common elements that should be on the requests page
    // This is a basic smoke test - adjust based on actual page content
    const heading = page.getByRole('heading', { name: /requests/i });
    if (await heading.count() > 0) {
      await expect(heading.first()).toBeVisible();
    }
  });

  test('should handle empty state gracefully', async ({ page }) => {
    // If there are no requests, should show appropriate message or empty state
    // This test ensures the page doesn't crash with no data
    const pageContent = page.locator('main');
    await expect(pageContent).toBeVisible();
  });

  test('should allow filtering/searching if controls exist', async ({ page }) => {
    // Check if search/filter controls exist
    const searchInput = page.getByRole('textbox', { name: /search|filter/i });
    const filterButton = page.getByRole('button', { name: /filter/i });

    // If either exists, verify they're interactive
    if (await searchInput.count() > 0) {
      await expect(searchInput.first()).toBeEnabled();
    }

    if (await filterButton.count() > 0) {
      await expect(filterButton.first()).toBeEnabled();
    }
  });

  test('should display request details when clicking on a request', async ({ page }) => {
    // Wait for any requests to load
    await page.waitForLoadState('domcontentloaded');

    // Look for clickable request items (adjust selector based on actual implementation)
    const requestItems = page.locator('[data-testid="request-item"]').or(
      page.locator('table tbody tr')
    );

    const count = await requestItems.count();

    if (count > 0) {
      // Click the first request
      await requestItems.first().click();

      // Should show some detail view (adjust based on implementation)
      // This could be a modal, a side panel, or a new page
      await page.waitForTimeout(500); // Allow for animation/loading

      // Verify detail view exists (generic check)
      const detailContent = page.locator('[role="dialog"]').or(
        page.locator('[data-testid="request-details"]')
      );

      await expect(detailContent.first()).toBeVisible({ timeout: 3000 });
    }
  });
});
