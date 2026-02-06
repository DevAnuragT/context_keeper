# Web Platform Integration Tests

This directory contains comprehensive integration tests for the MCP Context Engine web platform.

## Test Structure

### 1. Authentication Tests (`auth.integration.test.ts`)
Tests complete authentication workflows:
- **User Signup Flow**: Email/password registration, validation, duplicate detection
- **Email Verification**: Token validation, resend functionality
- **Login Flow**: Credential validation, session management, unverified email handling
- **OAuth Authentication**: GitHub/Google/Slack OAuth flows with CSRF protection
- **Password Reset**: Request and confirmation flows with security measures
- **Session Management**: Token storage, validation, expiration handling
- **Form Validation**: Email format, password strength requirements
- **Error Handling**: Network errors, malformed responses

### 2. Dashboard Tests (`dashboard.integration.test.ts`)
Tests project management and dashboard functionality:
- **Project Creation**: Validation, duplicate detection, authentication requirements
- **Project Listing**: Fetching, pagination, empty states
- **Project Details**: Comprehensive project information retrieval
- **Project Integrations**: Integration status, health monitoring
- **Project Metrics**: Statistics, activity tracking
- **Recent Activity**: Activity feed, filtering
- **Project Deletion**: Authorization, confirmation requirements
- **Real-time Updates**: Status polling, metric updates
- **Error Handling**: Network errors, session expiration
- **Data Formatting**: Date formatting, number formatting

### 3. Integration Wizard Tests (`integration-wizard.integration.test.ts`)
Tests the complete integration setup workflow:
- **Platform Selection**: Available platforms, selection UI
- **GitHub Integration**: OAuth flow, repository loading, configuration
- **Slack Integration**: OAuth flow, channel loading, DM inclusion
- **Discord Integration**: Bot installation, server/channel selection
- **Source Selection**: Multi-select, search/filter, validation
- **Configuration**: Platform-specific settings, validation
- **Initial Sync Monitoring**: Progress tracking, completion handling, error recovery
- **Wizard Navigation**: Step tracking, validation, state management
- **Error Handling**: OAuth errors, source loading failures
- **Wizard Completion**: Summary display, navigation options

### 4. End-to-End User Journey Tests (`e2e-user-journey.integration.test.ts`)
Tests complete user workflows:
- **Complete Onboarding**: Signup → Verify → Login → Dashboard
- **OAuth Onboarding**: OAuth signup → Dashboard
- **Project Setup**: Creation → Integration → Configuration → Sync
- **Multi-Platform Integration**: Adding multiple integrations
- **Context Querying**: Search, file context, decision retrieval
- **Error Recovery**: Failed sync recovery, session expiration
- **User Settings**: Profile updates, password changes
- **Team Collaboration**: Invitations, member management
- **Performance**: Caching, repeated requests

## Running Tests

### Run All Tests
```bash
npm test
```

### Run Specific Test Suite
```bash
npm test auth.integration.test.ts
npm test dashboard.integration.test.ts
npm test integration-wizard.integration.test.ts
npm test e2e-user-journey.integration.test.ts
```

### Run Tests in Watch Mode
```bash
npm run test:watch
```

### Run Tests with Coverage
```bash
npm test -- --coverage
```

## Test Coverage

The test suite covers:

### Authentication (auth.integration.test.ts)
- ✅ User signup with validation
- ✅ Email verification flow
- ✅ Login with email/password
- ✅ OAuth authentication (GitHub, Google, Slack)
- ✅ Password reset flow
- ✅ Session management
- ✅ CSRF protection
- ✅ Form validation
- ✅ Error handling

### Dashboard (dashboard.integration.test.ts)
- ✅ Project CRUD operations
- ✅ Project listing and pagination
- ✅ Integration status display
- ✅ Metrics and statistics
- ✅ Activity feed
- ✅ Real-time updates
- ✅ Authorization checks
- ✅ Error handling

### Integration Wizard (integration-wizard.integration.test.ts)
- ✅ Platform selection
- ✅ OAuth connection flows
- ✅ Source selection (repos, channels, servers)
- ✅ Configuration setup
- ✅ Initial sync monitoring
- ✅ Wizard navigation
- ✅ Error recovery
- ✅ Completion flow

### End-to-End Journeys (e2e-user-journey.integration.test.ts)
- ✅ Complete user onboarding
- ✅ Project creation and setup
- ✅ Multi-platform integration
- ✅ Context querying
- ✅ Error recovery flows
- ✅ User settings management
- ✅ Team collaboration
- ✅ Performance optimization

## Test Patterns

### Mocking
Tests use Vitest's mocking capabilities to mock:
- `fetch` API for HTTP requests
- `localStorage` for session storage
- `sessionStorage` for temporary data
- `crypto` for CSRF token generation

### Test Structure
Each test follows the Arrange-Act-Assert pattern:
```typescript
it('should do something', async () => {
  // Arrange: Set up test data and mocks
  mockFetch.mockResolvedValueOnce({ ... });
  
  // Act: Perform the action
  const response = await fetch('/api/endpoint', { ... });
  
  // Assert: Verify the results
  expect(response.ok).toBe(true);
});
```

### Async Testing
All API calls are properly awaited and tested:
```typescript
const response = await fetch('/api/endpoint');
const data = await response.json();
expect(data.success).toBe(true);
```

## Edge Cases Tested

### Authentication
- Invalid email formats
- Weak passwords
- Duplicate emails
- Expired tokens
- Invalid CSRF tokens
- Network errors

### Dashboard
- Empty project lists
- Unauthorized access
- Non-existent projects
- Session expiration
- Malformed responses

### Integration Wizard
- OAuth failures
- Source loading errors
- Configuration validation
- Sync failures
- Rate limiting

### End-to-End
- Session expiration during workflow
- Failed integration recovery
- Multi-step error handling
- Concurrent operations

## Best Practices

1. **Isolation**: Each test is independent and doesn't rely on other tests
2. **Cleanup**: `beforeEach` and `afterEach` hooks ensure clean state
3. **Mocking**: External dependencies are mocked for reliability
4. **Assertions**: Clear, specific assertions for each test case
5. **Error Cases**: Both success and failure paths are tested
6. **Real-world Scenarios**: Tests reflect actual user workflows

## Continuous Integration

These tests are designed to run in CI/CD pipelines:
- Fast execution (no real API calls)
- Deterministic results (mocked dependencies)
- Clear failure messages
- No external dependencies

## Future Enhancements

Potential additions to the test suite:
- Visual regression testing
- Performance benchmarking
- Accessibility testing
- Cross-browser compatibility
- Mobile responsiveness
- Load testing
- Security testing

## Contributing

When adding new features to the web platform:
1. Add corresponding integration tests
2. Follow existing test patterns
3. Test both success and error cases
4. Update this README with new test coverage
5. Ensure all tests pass before submitting PR

## Troubleshooting

### Tests Failing Locally
- Clear node_modules and reinstall: `rm -rf node_modules && npm install`
- Check Node version: `node --version` (requires >= 18.0.0)
- Run tests in isolation: `npm test -- auth.integration.test.ts`

### Mock Issues
- Ensure mocks are cleared in `afterEach`
- Check mock return values match expected API responses
- Verify mock call counts with `expect(mockFetch).toHaveBeenCalledTimes(n)`

### Async Issues
- Always await async operations
- Use `async/await` instead of `.then()`
- Check for unhandled promise rejections

## Resources

- [Vitest Documentation](https://vitest.dev/)
- [Testing Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)
- [Integration Testing Guide](https://martinfowler.com/bliki/IntegrationTest.html)
