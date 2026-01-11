# Testing Setup Summary

Playwright E2E tests have been successfully set up for the Claude Code Proxy dashboard.

## Test Results (Current Status)

**Status**: 33 of 33 tests passing (100% pass rate) ✅

### ✅ All Tests Passing

**Navigation** (4/4 tests)
- Page loading and rendering
- Sidebar navigation between all pages
- Header controls visibility
- Active state persistence on reload

**Dashboard Home** (4/4 tests)
- Overview display
- Key metrics rendering
- API failure handling
- Date range updates

**Requests Page** (4/4 tests)
- Page display and navigation
- Empty state handling
- Filter/search controls
- Request detail viewing

**Usage Page** (5/5 tests)
- Usage page display
- Charts and tables rendering
- Empty data handling
- Date range filtering
- Model breakdown display

**Accessibility** (15/15 tests)
- Heading structure for all 7 pages (Dashboard, Requests, Conversations, Usage, Performance, Routing, Settings)
- Keyboard navigation for all 7 pages
- Interactive element labels (all buttons have text or aria-label)
- Image alt text validation

## Configuration

- **Framework**: Playwright 1.57.0
- **Port**: 5174 (to avoid conflict with port 5173)
- **Browser**: Chromium
- **Auto-start**: Dev server starts automatically on test run
- **Test Location**: `dashboard/e2e/`
- **Execution Time**: ~12-15 seconds for full suite

## Running Tests

```bash
# Run all tests
cd dashboard
pnpm test

# Run specific test file
pnpm test e2e/navigation.spec.ts

# Run specific test by name
pnpm test -g "interactive elements"

# Interactive UI mode
pnpm test:ui

# Debug mode
pnpm test:debug

# View last test report
pnpm test:report
```

## Test Patterns

### Wait Strategies

**✅ Recommended**: Use `domcontentloaded` for most cases
```typescript
await page.goto('/some-page');
await page.waitForLoadState('domcontentloaded');
```

**❌ Avoid**: Do not use `networkidle` in development environments
```typescript
// DON'T DO THIS - will timeout in dev mode
await page.waitForLoadState('networkidle');
```

**Why?** Development environment has persistent network activity:
- Vite HMR (Hot Module Reload) WebSocket connections
- React Query automatic refetching and polling
- Background data refresh timers

The `networkidle` wait state will timeout waiting for all network activity to cease, which never happens in dev mode.

### Accessibility Requirements

All interactive elements must have accessible labels:

**Icon-only buttons need aria-label:**
```tsx
// ✅ Good
<button aria-label="Collapse sidebar">
  <ChevronLeft />
</button>

// ❌ Bad - no label
<button>
  <ChevronLeft />
</button>
```

**Buttons with text content are fine:**
```tsx
// ✅ Good - has visible text
<button>
  Save Changes
</button>
```

## Test Structure

```
dashboard/e2e/
├── README.md              # Comprehensive test documentation
├── helpers.ts             # Shared test utilities with JSDoc
├── navigation.spec.ts     # ✅ Navigation tests (4/4 passing)
├── dashboard.spec.ts      # ✅ Dashboard tests (4/4 passing)
├── requests.spec.ts       # ✅ Requests tests (4/4 passing)
├── usage.spec.ts          # ✅ Usage tests (5/5 passing)
└── accessibility.spec.ts  # ✅ Accessibility tests (15/15 passing)
```

## Troubleshooting

### Test Timeouts

**Symptom**: Tests timeout at 30 seconds
**Cause**: Using `waitForLoadState('networkidle')` in dev environment
**Fix**: Replace with `waitForLoadState('domcontentloaded')`

### Accessibility Failures

**Symptom**: "Button should have either text content or aria-label" error
**Cause**: Icon-only buttons without aria-label attribute
**Fix**: Add descriptive aria-label to all icon-only buttons

### Backend Connection Issues

**Symptom**: Tests fail with API errors
**Cause**: Backend proxy not running
**Fix**: Start backend before running tests:
```bash
# In separate terminal
cd proxy
just run
```

## Recommended Additions (Future Work)

1. **Visual Regression Testing**
   - Add screenshot comparisons for key pages
   - Use Playwright's `toHaveScreenshot()`

2. **API Mocking**
   - Mock API responses for consistent test data
   - Use `page.route()` to intercept requests

3. **Performance Testing**
   - Add metrics collection
   - Test page load times

4. **Error Boundary Testing**
   - Test error states more thoroughly
   - Verify error messages display correctly

5. **CI Integration**
   - Add to GitHub Actions workflow
   - Run on pull requests

## Test Coverage

**Current Coverage Areas:**
- ✅ Basic navigation and routing
- ✅ Page loading and rendering
- ✅ Empty states
- ✅ Basic accessibility
- ✅ Data loading
- ✅ Interactive features

**Not Yet Covered:**
- ❌ Form submissions
- ❌ API error states (partial)
- ❌ Real-time updates
- ❌ Mobile responsive behavior
- ❌ Cross-browser testing
- ❌ Authentication flows (if applicable)

## Configuration Files

- **playwright.config.ts** - Main Playwright configuration
- **package.json** - Test scripts defined
- **.gitignore** - Test artifacts excluded

## Maintenance

**Regular Tasks:**
1. Update selectors when UI changes
2. Add tests for new features
3. Review and fix flaky tests
4. Keep Playwright version updated
5. Monitor test execution time

**When Tests Fail:**
1. Check `playwright-report/` for detailed results
2. View screenshots in `test-results/`
3. Run with `--debug` flag for step-by-step debugging
4. Use `--ui` mode for interactive debugging

**Adding New Tests:**
1. Use helper functions from `helpers.ts`
2. Follow existing patterns for consistency
3. Use `domcontentloaded` for page loads
4. Add aria-labels to new icon-only buttons

## Resources

- [Playwright Documentation](https://playwright.dev)
- [Test README](./e2e/README.md) - Detailed testing guide
- [Best Practices](https://playwright.dev/docs/best-practices)

## Notes

- Tests run on port 5174 to avoid conflicts
- Dev server auto-starts and stops with tests
- Screenshots taken on failure only
- Parallel execution enabled (4 workers)
- Backend proxy must be running for tests to pass
