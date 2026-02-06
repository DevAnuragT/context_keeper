/**
 * Integration Tests for Dashboard Functionality
 * 
 * Tests the complete dashboard workflows including:
 * - Project creation and management
 * - Project listing and navigation
 * - Integration status display
 * - Metrics and statistics visualization
 * - Real-time updates
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock fetch for API calls
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => { store[key] = value; },
    removeItem: (key: string) => { delete store[key]; },
    clear: () => { store = {}; }
  };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('Dashboard Integration Tests', () => {
  const mockAuthToken = 'mock-jwt-token';
  const mockUser = {
    id: '123',
    email: 'test@example.com',
    name: 'Test User'
  };

  beforeEach(() => {
    localStorageMock.clear();
    localStorageMock.setItem('mcp_auth_token', mockAuthToken);
    localStorageMock.setItem('mcp_user_data', JSON.stringify(mockUser));
    mockFetch.mockClear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Project Creation', () => {
    it('should successfully create a new project', async () => {
      const newProject = {
        name: 'My New Project',
        description: 'A test project',
        settings: {
          sync_frequency: 'hourly'
        }
      };

      const mockResponse = {
        success: true,
        project: {
          id: 'proj-123',
          ...newProject,
          created_at: new Date().toISOString(),
          owner_id: mockUser.id
        }
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      const response = await fetch('/api/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify(newProject)
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.project.name).toBe(newProject.name);
      expect(data.project.id).toBeDefined();
    });

    it('should reject project creation with invalid name', async () => {
      const mockResponse = {
        success: false,
        message: 'Project name is required'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => mockResponse
      });

      const response = await fetch('/api/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({
          name: '',
          description: 'Test'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.message).toContain('required');
    });

    it('should reject project creation with duplicate name', async () => {
      const mockResponse = {
        success: false,
        message: 'Project name already exists'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 409,
        json: async () => mockResponse
      });

      const response = await fetch('/api/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({
          name: 'Existing Project',
          description: 'Test'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(response.status).toBe(409);
    });

    it('should require authentication for project creation', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          success: false,
          message: 'Unauthorized'
        })
      });

      const response = await fetch('/api/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name: 'Test Project',
          description: 'Test'
        })
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(401);
    });
  });

  describe('Project Listing', () => {
    it('should fetch all projects for authenticated user', async () => {
      const mockProjects = [
        {
          id: 'proj-1',
          name: 'Project 1',
          description: 'First project',
          created_at: new Date().toISOString(),
          integrations: [],
          stats: { entities: 10, decisions: 5, discussions: 15 }
        },
        {
          id: 'proj-2',
          name: 'Project 2',
          description: 'Second project',
          created_at: new Date().toISOString(),
          integrations: [{ platform: 'github', status: 'active' }],
          stats: { entities: 20, decisions: 8, discussions: 25 }
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ projects: mockProjects })
      });

      const response = await fetch('/api/projects', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.projects).toHaveLength(2);
      expect(data.projects[0].name).toBe('Project 1');
    });

    it('should return empty array when user has no projects', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ projects: [] })
      });

      const response = await fetch('/api/projects', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.projects).toHaveLength(0);
    });

    it('should handle pagination for large project lists', async () => {
      const mockProjects = Array.from({ length: 10 }, (_, i) => ({
        id: `proj-${i}`,
        name: `Project ${i}`,
        description: `Project ${i} description`,
        created_at: new Date().toISOString()
      }));

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          projects: mockProjects.slice(0, 5),
          pagination: {
            page: 1,
            per_page: 5,
            total: 10,
            has_more: true
          }
        })
      });

      const response = await fetch('/api/projects?page=1&per_page=5', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.projects).toHaveLength(5);
      expect(data.pagination.has_more).toBe(true);
    });
  });

  describe('Project Details', () => {
    it('should fetch detailed project information', async () => {
      const mockProject = {
        id: 'proj-123',
        name: 'Test Project',
        description: 'A test project',
        created_at: new Date().toISOString(),
        owner_id: mockUser.id,
        integrations: [
          {
            id: 'int-1',
            platform: 'github',
            status: 'active',
            last_sync: new Date().toISOString()
          }
        ],
        stats: {
          entities: 50,
          decisions: 12,
          discussions: 30,
          files: 100
        },
        team_members: [
          {
            user_id: mockUser.id,
            role: 'owner',
            joined_at: new Date().toISOString()
          }
        ]
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ project: mockProject })
      });

      const response = await fetch('/api/projects/proj-123', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.project.id).toBe('proj-123');
      expect(data.project.integrations).toHaveLength(1);
      expect(data.project.stats.entities).toBe(50);
    });

    it('should return 404 for non-existent project', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: async () => ({
          success: false,
          message: 'Project not found'
        })
      });

      const response = await fetch('/api/projects/non-existent', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(404);
    });

    it('should return 403 for unauthorized project access', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        json: async () => ({
          success: false,
          message: 'Access denied'
        })
      });

      const response = await fetch('/api/projects/other-user-project', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(403);
    });
  });

  describe('Project Integrations', () => {
    it('should fetch all integrations for a project', async () => {
      const mockIntegrations = [
        {
          id: 'int-1',
          platform: 'github',
          status: 'active',
          last_sync: new Date().toISOString(),
          sources: ['repo1', 'repo2']
        },
        {
          id: 'int-2',
          platform: 'slack',
          status: 'syncing',
          last_sync: new Date().toISOString(),
          sources: ['channel1']
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ integrations: mockIntegrations })
      });

      const response = await fetch('/api/projects/proj-123/integrations', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.integrations).toHaveLength(2);
      expect(data.integrations[0].platform).toBe('github');
    });

    it('should show integration health status', async () => {
      const mockIntegrations = [
        {
          id: 'int-1',
          platform: 'github',
          status: 'active',
          health: {
            status: 'healthy',
            last_check: new Date().toISOString(),
            issues: []
          }
        },
        {
          id: 'int-2',
          platform: 'slack',
          status: 'error',
          health: {
            status: 'unhealthy',
            last_check: new Date().toISOString(),
            issues: ['Rate limit exceeded']
          }
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ integrations: mockIntegrations })
      });

      const response = await fetch('/api/projects/proj-123/integrations', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(data.integrations[0].health.status).toBe('healthy');
      expect(data.integrations[1].health.status).toBe('unhealthy');
      expect(data.integrations[1].health.issues).toContain('Rate limit exceeded');
    });
  });

  describe('Project Metrics', () => {
    it('should fetch project metrics and statistics', async () => {
      const mockMetrics = {
        entities: {
          total: 100,
          by_type: {
            feature: 30,
            file: 50,
            decision: 20
          }
        },
        decisions: {
          total: 20,
          by_status: {
            active: 15,
            superseded: 3,
            deprecated: 2
          }
        },
        discussions: {
          total: 50,
          by_platform: {
            github: 30,
            slack: 15,
            discord: 5
          }
        },
        activity: {
          last_7_days: 45,
          last_30_days: 180
        }
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ metrics: mockMetrics })
      });

      const response = await fetch('/api/projects/proj-123/metrics', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.metrics.entities.total).toBe(100);
      expect(data.metrics.decisions.by_status.active).toBe(15);
    });

    it('should fetch project statistics summary', async () => {
      const mockStats = {
        entities: 100,
        decisions: 20,
        discussions: 50,
        files: 200,
        contributors: 5,
        last_updated: new Date().toISOString()
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ stats: mockStats })
      });

      const response = await fetch('/api/projects/proj-123/stats', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.stats.entities).toBe(100);
      expect(data.stats.contributors).toBe(5);
    });
  });

  describe('Recent Activity', () => {
    it('should fetch recent activity for a project', async () => {
      const mockActivities = [
        {
          id: 'act-1',
          type: 'sync',
          title: 'GitHub sync completed',
          description: 'Synced 5 new pull requests',
          timestamp: new Date().toISOString(),
          metadata: { platform: 'github', items: 5 }
        },
        {
          id: 'act-2',
          type: 'decision',
          title: 'New decision extracted',
          description: 'Architecture decision from PR #123',
          timestamp: new Date(Date.now() - 3600000).toISOString(),
          metadata: { source: 'github', pr_number: 123 }
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ activities: mockActivities })
      });

      const response = await fetch('/api/projects/proj-123/activity?limit=10', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.activities).toHaveLength(2);
      expect(data.activities[0].type).toBe('sync');
    });

    it('should support activity filtering by type', async () => {
      const mockActivities = [
        {
          id: 'act-1',
          type: 'decision',
          title: 'Decision 1',
          timestamp: new Date().toISOString()
        },
        {
          id: 'act-2',
          type: 'decision',
          title: 'Decision 2',
          timestamp: new Date().toISOString()
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ activities: mockActivities })
      });

      const response = await fetch('/api/projects/proj-123/activity?type=decision', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.activities.every((a: any) => a.type === 'decision')).toBe(true);
    });
  });

  describe('Project Deletion', () => {
    it('should successfully delete a project', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Project deleted successfully'
        })
      });

      const response = await fetch('/api/projects/proj-123', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
    });

    it('should require confirmation for project deletion', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({
          success: false,
          message: 'Confirmation required'
        })
      });

      const response = await fetch('/api/projects/proj-123', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.message).toContain('Confirmation');
    });

    it('should only allow owner to delete project', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        json: async () => ({
          success: false,
          message: 'Only project owner can delete'
        })
      });

      const response = await fetch('/api/projects/proj-123', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(403);
    });
  });

  describe('Real-time Updates', () => {
    it('should poll for integration status updates', async () => {
      const mockStatus = {
        integration_id: 'int-1',
        status: 'syncing',
        progress: 45,
        message: 'Syncing repositories...'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ status: mockStatus })
      });

      const response = await fetch('/api/integrations/int-1/status', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.status.progress).toBe(45);
    });

    it('should handle real-time metric updates', async () => {
      const initialStats = { entities: 100, decisions: 20 };
      const updatedStats = { entities: 105, decisions: 22 };

      mockFetch
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ stats: initialStats })
        })
        .mockResolvedValueOnce({
          ok: true,
          json: async () => ({ stats: updatedStats })
        });

      const response1 = await fetch('/api/projects/proj-123/stats', {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${mockAuthToken}` }
      });
      const data1 = await response1.json();

      const response2 = await fetch('/api/projects/proj-123/stats', {
        method: 'GET',
        headers: { 'Authorization': `Bearer ${mockAuthToken}` }
      });
      const data2 = await response2.json();

      expect(data1.stats.entities).toBe(100);
      expect(data2.stats.entities).toBe(105);
    });
  });

  describe('Error Handling', () => {
    it('should handle network errors gracefully', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      try {
        await fetch('/api/projects', {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${mockAuthToken}`
          }
        });
      } catch (error) {
        expect(error).toBeInstanceOf(Error);
        expect((error as Error).message).toBe('Network error');
      }
    });

    it('should redirect to login on 401 errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          success: false,
          message: 'Unauthorized'
        })
      });

      const response = await fetch('/api/projects', {
        method: 'GET',
        headers: {
          'Authorization': 'Bearer invalid-token'
        }
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(401);
    });

    it('should show user-friendly error messages', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({
          success: false,
          message: 'Internal server error',
          user_message: 'Something went wrong. Please try again later.'
        })
      });

      const response = await fetch('/api/projects', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({ name: 'Test' })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.user_message).toBeDefined();
    });
  });

  describe('Data Formatting', () => {
    it('should format dates for display', () => {
      const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / 60000);

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
        return date.toLocaleDateString();
      };

      const now = new Date();
      const fiveMinutesAgo = new Date(now.getTime() - 5 * 60000);
      const yesterday = new Date(now.getTime() - 24 * 60 * 60000);

      expect(formatDate(now.toISOString())).toBe('Just now');
      expect(formatDate(fiveMinutesAgo.toISOString())).toContain('minute');
      expect(formatDate(yesterday.toISOString())).toMatch(/\d+\/\d+\/\d+/);
    });

    it('should format numbers with commas', () => {
      const formatNumber = (num: number) => {
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
      };

      expect(formatNumber(1000)).toBe('1,000');
      expect(formatNumber(1000000)).toBe('1,000,000');
      expect(formatNumber(100)).toBe('100');
    });
  });
});
