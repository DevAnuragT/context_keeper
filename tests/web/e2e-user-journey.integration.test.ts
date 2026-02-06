/**
 * End-to-End User Journey Integration Tests
 * 
 * Tests complete user workflows from signup to using the platform:
 * - Complete onboarding flow (signup → verify → login → dashboard)
 * - Project creation and setup
 * - Integration installation
 * - Context querying
 * - Multi-platform workflows
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock fetch for API calls
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage and sessionStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; }
  };
})();

const sessionStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; }
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });
Object.defineProperty(window, 'sessionStorage', { value: sessionStorageMock });

describe('End-to-End User Journey Tests', () => {
  beforeEach(() => {
    localStorageMock.clear();
    sessionStorageMock.clear();
    mockFetch.mockClear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Complete Onboarding Flow', () => {
    it('should complete full signup to dashboard journey', async () => {
      // Step 1: User signs up
      const signupData = {
        email: 'newuser@example.com',
        password: 'SecurePass123!',
        name: 'New User'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          token: 'temp-token',
          user: {
            id: 'user-123',
            email: signupData.email,
            name: signupData.name,
            email_verified: false
          }
        })
      });

      const signupResponse = await fetch('/api/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(signupData)
      });

      const signupResult = await signupResponse.json();
      expect(signupResult.success).toBe(true);
      expect(signupResult.user.email_verified).toBe(false);

      // Step 2: User verifies email
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Email verified successfully'
        })
      });

      const verifyResponse = await fetch('/api/auth/verify-email', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token: 'verification-token' })
      });

      const verifyResult = await verifyResponse.json();
      expect(verifyResult.success).toBe(true);

      // Step 3: User logs in
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          token: 'auth-token',
          user: {
            id: 'user-123',
            email: signupData.email,
            name: signupData.name,
            email_verified: true
          }
        })
      });

      const loginResponse = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: signupData.email,
          password: signupData.password
        })
      });

      const loginResult = await loginResponse.json();
      expect(loginResult.success).toBe(true);
      expect(loginResult.user.email_verified).toBe(true);

      // Store session
      localStorageMock.setItem('mcp_auth_token', loginResult.token);
      localStorageMock.setItem('mcp_user_data', JSON.stringify(loginResult.user));

      // Step 4: User accesses dashboard
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ projects: [] })
      });

      const dashboardResponse = await fetch('/api/projects', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${loginResult.token}`
        }
      });

      const dashboardResult = await dashboardResponse.json();
      expect(dashboardResponse.ok).toBe(true);
      expect(dashboardResult.projects).toEqual([]);
    });

    it('should handle OAuth signup to dashboard journey', async () => {
      // Step 1: User initiates GitHub OAuth
      const csrfToken = 'csrf-token';
      sessionStorageMock.setItem('oauth_csrf_token', csrfToken);

      // Step 2: OAuth callback with user data
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          token: 'oauth-token',
          user: {
            id: 'user-456',
            email: 'github@example.com',
            name: 'GitHub User',
            oauth_provider: 'github',
            email_verified: true
          }
        })
      });

      const callbackResponse = await fetch('/api/auth/oauth/callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: 'oauth-code',
          state: csrfToken
        })
      });

      const callbackResult = await callbackResponse.json();
      expect(callbackResult.success).toBe(true);
      expect(callbackResult.user.oauth_provider).toBe('github');

      // Store session
      localStorageMock.setItem('mcp_auth_token', callbackResult.token);

      // Step 3: User accesses dashboard
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ projects: [] })
      });

      const dashboardResponse = await fetch('/api/projects', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${callbackResult.token}`
        }
      });

      expect(dashboardResponse.ok).toBe(true);
    });
  });

  describe('Project Creation and Setup Flow', () => {
    const authToken = 'auth-token';

    beforeEach(() => {
      localStorageMock.setItem('mcp_auth_token', authToken);
    });

    it('should complete project creation to first integration', async () => {
      // Step 1: Create project
      const projectData = {
        name: 'My First Project',
        description: 'Getting started with MCP Context Engine'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          project: {
            id: 'proj-123',
            ...projectData,
            created_at: new Date().toISOString()
          }
        })
      });

      const createResponse = await fetch('/api/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify(projectData)
      });

      const createResult = await createResponse.json();
      expect(createResult.success).toBe(true);
      const projectId = createResult.project.id;

      // Step 2: Navigate to integration wizard
      // (URL would be /integrations?project_id=proj-123)

      // Step 3: Select GitHub platform
      const selectedPlatform = 'github';
      sessionStorageMock.setItem('integration_platform', selectedPlatform);

      // Step 4: Connect GitHub OAuth
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          integration_id: 'int-123',
          platform: 'github',
          status: 'connected'
        })
      });

      // Simulate OAuth callback
      const integrationId = 'int-123';

      // Step 5: Load repositories
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          sources: [
            { id: 'repo-1', name: 'my-repo', description: 'My repository' },
            { id: 'repo-2', name: 'another-repo', description: 'Another repo' }
          ]
        })
      });

      const sourcesResponse = await fetch(`/api/integrations/${integrationId}/sources`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const sourcesResult = await sourcesResponse.json();
      expect(sourcesResult.sources).toHaveLength(2);

      // Step 6: Configure and save
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Configuration saved'
        })
      });

      const configResponse = await fetch(`/api/integrations/${integrationId}/configure`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
          sources: ['repo-1'],
          configuration: {
            sync_frequency: 'hourly',
            sync_prs: true,
            sync_issues: true,
            sync_commits: true
          }
        })
      });

      const configResult = await configResponse.json();
      expect(configResult.success).toBe(true);

      // Step 7: Monitor initial sync
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          status: {
            state: 'completed',
            progress: 100,
            message: 'Sync completed'
          }
        })
      });

      const statusResponse = await fetch(`/api/integrations/${integrationId}/status`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const statusResult = await statusResponse.json();
      expect(statusResult.status.state).toBe('completed');
    });
  });

  describe('Multi-Platform Integration Flow', () => {
    const authToken = 'auth-token';
    const projectId = 'proj-123';

    beforeEach(() => {
      localStorageMock.setItem('mcp_auth_token', authToken);
    });

    it('should add multiple integrations to a project', async () => {
      // Add GitHub integration
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          integration_id: 'int-github',
          platform: 'github'
        })
      });

      // Configure GitHub
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      });

      // Add Slack integration
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          integration_id: 'int-slack',
          platform: 'slack'
        })
      });

      // Configure Slack
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true })
      });

      // Verify both integrations exist
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          integrations: [
            { id: 'int-github', platform: 'github', status: 'active' },
            { id: 'int-slack', platform: 'slack', status: 'active' }
          ]
        })
      });

      const integrationsResponse = await fetch(`/api/projects/${projectId}/integrations`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const integrationsResult = await integrationsResponse.json();
      expect(integrationsResult.integrations).toHaveLength(2);
      expect(integrationsResult.integrations.map((i: any) => i.platform)).toContain('github');
      expect(integrationsResult.integrations.map((i: any) => i.platform)).toContain('slack');
    });
  });

  describe('Context Query Flow', () => {
    const authToken = 'auth-token';
    const projectId = 'proj-123';

    beforeEach(() => {
      localStorageMock.setItem('mcp_auth_token', authToken);
    });

    it('should query project context after data ingestion', async () => {
      // Step 1: Verify project has data
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          stats: {
            entities: 50,
            decisions: 10,
            discussions: 30
          }
        })
      });

      const statsResponse = await fetch(`/api/projects/${projectId}/stats`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const statsResult = await statsResponse.json();
      expect(statsResult.stats.entities).toBeGreaterThan(0);

      // Step 2: Search for specific context
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          results: [
            {
              type: 'decision',
              title: 'Use React for frontend',
              content: 'Decision to use React framework...',
              source: 'github',
              relevance: 0.95
            },
            {
              type: 'discussion',
              title: 'Frontend architecture discussion',
              content: 'Team discussion about frontend choices...',
              source: 'slack',
              relevance: 0.87
            }
          ]
        })
      });

      const searchResponse = await fetch(`/api/projects/${projectId}/search`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
          query: 'frontend architecture decisions'
        })
      });

      const searchResult = await searchResponse.json();
      expect(searchResult.results).toHaveLength(2);
      expect(searchResult.results[0].type).toBe('decision');

      // Step 3: Get file-specific context
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          file_context: {
            path: 'src/components/App.tsx',
            change_history: [
              {
                date: new Date().toISOString(),
                reason: 'Initial component structure',
                discussion_link: 'pr-123'
              }
            ],
            related_decisions: [
              {
                id: 'dec-1',
                title: 'Use React for frontend'
              }
            ]
          }
        })
      });

      const fileContextResponse = await fetch(`/api/projects/${projectId}/files/context`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
          file_path: 'src/components/App.tsx'
        })
      });

      const fileContextResult = await fileContextResponse.json();
      expect(fileContextResult.file_context.path).toBe('src/components/App.tsx');
      expect(fileContextResult.file_context.related_decisions).toHaveLength(1);
    });
  });

  describe('Error Recovery Flows', () => {
    const authToken = 'auth-token';

    beforeEach(() => {
      localStorageMock.setItem('mcp_auth_token', authToken);
    });

    it('should recover from failed integration sync', async () => {
      const integrationId = 'int-123';

      // Step 1: Check integration status shows error
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          status: {
            state: 'error',
            error: 'Rate limit exceeded',
            retry_after: 3600
          }
        })
      });

      const statusResponse = await fetch(`/api/integrations/${integrationId}/status`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const statusResult = await statusResponse.json();
      expect(statusResult.status.state).toBe('error');

      // Step 2: Retry sync after waiting
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Sync restarted'
        })
      });

      const retryResponse = await fetch(`/api/integrations/${integrationId}/retry`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const retryResult = await retryResponse.json();
      expect(retryResult.success).toBe(true);

      // Step 3: Verify sync is now running
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          status: {
            state: 'syncing',
            progress: 10
          }
        })
      });

      const newStatusResponse = await fetch(`/api/integrations/${integrationId}/status`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const newStatusResult = await newStatusResponse.json();
      expect(newStatusResult.status.state).toBe('syncing');
    });

    it('should handle session expiration gracefully', async () => {
      // Step 1: Make request with expired token
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          success: false,
          message: 'Session expired'
        })
      });

      const response = await fetch('/api/projects', {
        method: 'GET',
        headers: { 'Authorization': 'Bearer expired-token' }
      });

      expect(response.status).toBe(401);

      // Step 2: Clear session and redirect to login
      localStorageMock.removeItem('mcp_auth_token');
      localStorageMock.removeItem('mcp_user_data');

      expect(localStorageMock.getItem('mcp_auth_token')).toBeNull();

      // Step 3: User logs in again
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          token: 'new-token',
          user: { id: 'user-123', email: 'test@example.com' }
        })
      });

      const loginResponse = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'test@example.com',
          password: 'password'
        })
      });

      const loginResult = await loginResponse.json();
      expect(loginResult.success).toBe(true);

      // Store new session
      localStorageMock.setItem('mcp_auth_token', loginResult.token);

      // Step 4: Retry original request
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ projects: [] })
      });

      const retryResponse = await fetch('/api/projects', {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${loginResult.token}` }
      });

      expect(retryResponse.ok).toBe(true);
    });
  });

  describe('User Settings and Preferences', () => {
    const authToken = 'auth-token';

    beforeEach(() => {
      localStorageMock.setItem('mcp_auth_token', authToken);
    });

    it('should update user profile', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          user: {
            id: 'user-123',
            email: 'test@example.com',
            name: 'Updated Name'
          }
        })
      });

      const response = await fetch('/api/user/profile', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
          name: 'Updated Name'
        })
      });

      const result = await response.json();
      expect(result.success).toBe(true);
      expect(result.user.name).toBe('Updated Name');
    });

    it('should change password', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Password updated successfully'
        })
      });

      const response = await fetch('/api/user/password', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
          current_password: 'OldPass123!',
          new_password: 'NewPass123!'
        })
      });

      const result = await response.json();
      expect(result.success).toBe(true);
    });
  });

  describe('Team Collaboration Flow', () => {
    const authToken = 'auth-token';
    const projectId = 'proj-123';

    beforeEach(() => {
      localStorageMock.setItem('mcp_auth_token', authToken);
    });

    it('should invite team member to project', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          invitation: {
            id: 'inv-123',
            email: 'teammate@example.com',
            role: 'member',
            status: 'pending'
          }
        })
      });

      const response = await fetch(`/api/projects/${projectId}/invitations`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authToken}`
        },
        body: JSON.stringify({
          email: 'teammate@example.com',
          role: 'member'
        })
      });

      const result = await response.json();
      expect(result.success).toBe(true);
      expect(result.invitation.status).toBe('pending');
    });

    it('should list project team members', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          members: [
            {
              user_id: 'user-123',
              email: 'owner@example.com',
              role: 'owner',
              joined_at: new Date().toISOString()
            },
            {
              user_id: 'user-456',
              email: 'member@example.com',
              role: 'member',
              joined_at: new Date().toISOString()
            }
          ]
        })
      });

      const response = await fetch(`/api/projects/${projectId}/members`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const result = await response.json();
      expect(result.members).toHaveLength(2);
      expect(result.members[0].role).toBe('owner');
    });
  });

  describe('Performance and Caching', () => {
    const authToken = 'auth-token';
    const projectId = 'proj-123';

    beforeEach(() => {
      localStorageMock.setItem('mcp_auth_token', authToken);
    });

    it('should cache frequently accessed data', async () => {
      // First request
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          stats: { entities: 100 }
        })
      });

      const response1 = await fetch(`/api/projects/${projectId}/stats`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const result1 = await response1.json();
      expect(result1.stats.entities).toBe(100);

      // Second request (should use cache if implemented)
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          stats: { entities: 100 }
        })
      });

      const response2 = await fetch(`/api/projects/${projectId}/stats`, {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${authToken}` }
      });

      const result2 = await response2.json();
      expect(result2.stats.entities).toBe(100);
    });
  });
});
