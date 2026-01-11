import { Page, expect } from '@playwright/test';

/**
 * Helper functions for E2E tests
 */

/**
 * Wait for the dashboard to finish loading.
 *
 * Uses 'domcontentloaded' instead of 'networkidle' because the development environment
 * has persistent network activity (Vite HMR, React Query polling) that prevents networkidle.
 * This ensures the DOM is ready for interaction without waiting for background requests.
 *
 * @param page - Playwright Page object
 */
export async function waitForDashboardLoad(page: Page) {
  await page.waitForLoadState('domcontentloaded');
  // Wait for main content to be visible
  await expect(page.locator('main')).toBeVisible();
}

/**
 * Navigate to a specific page using the sidebar
 */
export async function navigateToPage(page: Page, pageName: string) {
  await page.getByRole('link', { name: new RegExp(pageName, 'i') }).click();
  await waitForDashboardLoad(page);
}

/**
 * Wait for a specific API response before proceeding.
 * Useful for ensuring data has loaded before making assertions.
 *
 * @param page - Playwright Page object
 * @param endpoint - API endpoint pattern to wait for (e.g., '/api/v2/requests')
 * @param timeout - Maximum time to wait in milliseconds (default: 5000)
 *
 * @example
 * await waitForAPIResponse(page, '/api/v2/requests/summary');
 */
export async function waitForAPIResponse(
  page: Page,
  endpoint: string,
  timeout = 5000
) {
  try {
    await page.waitForResponse(
      (response) => response.url().includes(endpoint) && response.status() === 200,
      { timeout }
    );
  } catch {
    // Response may have already completed before we started waiting
    // This is acceptable - we just want to ensure it's not still loading
  }
}

/**
 * Check if the page is showing an error state
 */
export async function hasErrorState(page: Page): Promise<boolean> {
  const errorElements = page.getByText(/error|failed|something went wrong/i);
  return (await errorElements.count()) > 0;
}

/**
 * Check if the page is showing a loading state
 */
export async function hasLoadingState(page: Page): Promise<boolean> {
  const loadingElements = page.getByText(/loading|spinner/i).or(
    page.locator('[data-testid="loading"]')
  );
  return (await loadingElements.count()) > 0;
}

/**
 * Wait for any loading spinners to disappear
 */
export async function waitForLoadingComplete(page: Page, timeout = 5000) {
  const loadingElements = page.getByText(/loading/i).or(
    page.locator('[data-testid="loading"]')
  );

  try {
    await loadingElements.first().waitFor({ state: 'hidden', timeout });
  } catch {
    // No loading elements found or already hidden
  }
}

/**
 * Take a screenshot with a descriptive name
 */
export async function takeDebugScreenshot(page: Page, name: string) {
  await page.screenshot({ path: `screenshots/${name}.png`, fullPage: true });
}

/**
 * Get the current date range from the global date picker
 */
export async function getCurrentDateRange(page: Page): Promise<string | null> {
  const datePicker = page.locator('[data-testid="date-picker"]').or(
    page.getByText(/today|week|date/i).first()
  );

  if ((await datePicker.count()) > 0) {
    return await datePicker.first().textContent();
  }

  return null;
}

/**
 * Assert that the API is responding
 */
export async function assertAPIHealthy(page: Page) {
  // Make a request to the health endpoint
  const response = await page.request.get('/health');
  expect(response.ok()).toBeTruthy();
}

/**
 * Wait for conversations to load and verify data is displayed.
 * This ensures the API actually returns data, not just that the DOM loaded.
 *
 * @param page - Playwright Page object
 * @param minConversations - Minimum number of conversations expected (default: 1)
 */
export async function waitForConversationsLoad(page: Page, minConversations = 1) {
  // Wait for API response
  await waitForAPIResponse(page, '/api/v2/conversations');

  // Wait for DOM to be ready
  await page.waitForLoadState('domcontentloaded');

  // Verify conversation count is displayed
  const countText = page.locator('text=/\\d+ conversation/');
  await expect(countText).toBeVisible({ timeout: 5000 });

  // If we expect conversations, verify at least one conversation button exists
  if (minConversations > 0) {
    const conversationButtons = page.locator('main button').filter({
      hasText: /.ago$/ // Conversations end with "Xm ago", "Xh ago", etc.
    });
    await expect(conversationButtons.first()).toBeVisible({ timeout: 5000 });
  }
}
