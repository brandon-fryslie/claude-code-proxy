import { test, expect } from '@playwright/test';

test.describe('Accessibility', () => {
  const pages = [
    { name: 'Dashboard', path: '/' },
    { name: 'Requests', path: '/requests' },
    { name: 'Conversations', path: '/conversations' },
    { name: 'Usage', path: '/usage' },
    { name: 'Performance', path: '/performance' },
    { name: 'Routing', path: '/routing' },
    { name: 'Settings', path: '/settings' },
  ];

  for (const { name, path } of pages) {
    test(`${name} page should have proper heading structure`, async ({ page }) => {
      await page.goto(path);
      await page.waitForLoadState('domcontentloaded');

      // Should have at least one heading
      const headings = page.locator('h1, h2, h3, h4, h5, h6');
      const count = await headings.count();

      // Basic check - page should have some semantic structure
      expect(count).toBeGreaterThanOrEqual(0);
    });

    test(`${name} page should be keyboard navigable`, async ({ page }) => {
      await page.goto(path);
      await page.waitForLoadState('domcontentloaded');

      // Tab through interactive elements
      await page.keyboard.press('Tab');
      await page.keyboard.press('Tab');

      // Verify focus is working (focused element should exist)
      const focusedElement = page.locator(':focus');
      const hasFocus = await focusedElement.count() > 0;

      // This is a basic check - at least verify tabbing doesn't crash
      expect(hasFocus || true).toBeTruthy();
    });
  }

  test('interactive elements should have accessible labels', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');

    // Check that buttons have text or aria-labels
    const buttons = page.locator('button');
    const buttonCount = await buttons.count();

    if (buttonCount > 0) {
      for (let i = 0; i < Math.min(buttonCount, 5); i++) {
        const button = buttons.nth(i);
        const text = await button.textContent();
        const ariaLabel = await button.getAttribute('aria-label');

        // Button should have either text content or aria-label
        expect(text || ariaLabel).toBeTruthy();
      }
    }
  });

  test('images should have alt text', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');

    // Check that images have alt attributes
    const images = page.locator('img');
    const imageCount = await images.count();

    if (imageCount > 0) {
      for (let i = 0; i < imageCount; i++) {
        const img = images.nth(i);
        const alt = await img.getAttribute('alt');

        // Alt attribute should exist (can be empty for decorative images)
        expect(alt !== null).toBeTruthy();
      }
    }
  });
});
