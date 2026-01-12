# Dashboard E2E Tests

End-to-end tests for the Claude Code Proxy dashboard using Playwright.

## Structure

```
e2e/
├── README.md              # This file
├── helpers.ts             # Shared test utilities
├── navigation.spec.ts     # Navigation and routing tests
├── dashboard.spec.ts      # Dashboard home page tests
├── requests.spec.ts       # Requests page tests
├── usage.spec.ts          # Usage/analytics page tests
├── accessibility.spec.ts  # Accessibility compliance tests
└── *.spec.ts             # Additional test files
```

## Running Tests

```bash
# Run all tests
pnpm test

# Run tests in UI mode (interactive)
pnpm test:ui

# Run tests in debug mode
pnpm test:debug

# View last test report
pnpm test:report

# Run specific test file
pnpm test e2e/navigation.spec.ts

# Run tests matching a pattern
pnpm test --grep "navigation"
```

## Writing Tests

### Basic Test Structure

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/your-page');
  });

  test('should do something', async ({ page }) => {
    // Your test code
    await expect(page.locator('selector')).toBeVisible();
  });
});
```

### Using Helpers

```typescript
import { waitForDashboardLoad, navigateToPage } from './helpers';

test('example', async ({ page }) => {
  await page.goto('/');
  await waitForDashboardLoad(page);
  await navigateToPage(page, 'requests');
});
```

## Test Strategy

### Coverage Areas

1. **Navigation** - Routing, sidebar, page transitions
2. **Core Pages** - Dashboard, Requests, Usage, Conversations
3. **User Interactions** - Clicks, filters, date selection
4. **Error Handling** - API failures, network issues
5. **Accessibility** - Keyboard navigation, screen readers
6. **Performance** - Page load times, rendering

### What to Test

✅ **Do Test:**
- Critical user workflows
- Navigation between pages
- Data loading and display
- Error states and fallbacks
- Accessibility compliance
- Keyboard navigation

❌ **Don't Test:**
- Implementation details
- Component internals (use unit tests)
- Styling specifics (use visual regression tests)
- Third-party library functionality

## Configuration

See `playwright.config.ts` for configuration:

- **Base URL:** `http://localhost:5173`
- **Browser:** Chromium (others commented out)
- **Retries:** 2 on CI, 0 locally
- **Screenshots:** On failure only
- **Traces:** On first retry

## CI Integration

Tests run automatically on CI with:
- Parallel execution disabled (workers: 1)
- Retries enabled (2 attempts)
- Fail-fast on `.only` tests

## Debugging

### View Test Report

```bash
pnpm test:report
```

### Debug Mode

```bash
pnpm test:debug
```

This opens Playwright Inspector for step-by-step debugging.

### Take Screenshots

```typescript
await page.screenshot({ path: 'debug.png' });
```

### Console Logs

```typescript
page.on('console', msg => console.log(msg.text()));
```

## Best Practices

1. **Use Semantic Selectors**
   - Prefer `getByRole()`, `getByLabel()`, `getByText()`
   - Avoid CSS selectors when possible

2. **Wait Properly**
   - Use `waitForLoadState('networkidle')`
   - Use auto-waiting with `expect()`
   - Avoid fixed `setTimeout()` when possible

3. **Test Data Independence**
   - Don't depend on specific data
   - Test empty states
   - Handle variable data gracefully

4. **Assertions**
   - Be specific but flexible
   - Test behavior, not implementation
   - Use semantic matchers

5. **Test Isolation**
   - Each test should be independent
   - Use `beforeEach` for setup
   - Clean up after tests if needed

## Troubleshooting

### Tests timing out
- Increase timeout in `playwright.config.ts`
- Check if dev server is running
- Verify network conditions

### Flaky tests
- Add proper waits (`waitForLoadState`)
- Check for race conditions
- Use `test.setTimeout()` for slow tests

### Element not found
- Verify selector is correct
- Check if element is in viewport
- Use `page.pause()` to debug interactively

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Debugging Guide](https://playwright.dev/docs/debug)
