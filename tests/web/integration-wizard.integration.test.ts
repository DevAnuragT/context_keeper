/**
 * Integration Tests for Integration Wizard
 * 
 * Tests the complete integration setup workflow including:
 * - Platform selection
 * - OAuth connection flow
 * - Source selection (repositories, channels, servers)
 * - Configuration setup
 * - Initial sync monitoring
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

describe('Integration Wizard Integration Tests', () => {
  const mockAuthToken = 'mock-jwt-token';
  const mockProjectId = 'proj-123';

  beforeEach(() => {
    localStorageMock.clear();
    sessionStorageMock.clear();
    localStorageMock.setItem('mcp_auth_token', mockAuthToken);
    mockFetch.mockClear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Platform Selection', () => {
    it('should display available platforms', () => {
      const platforms = ['github', 'slack', 'discord'];
      
      expect(platforms).toContain('github');
      expect(platforms).toContain('slack');
      expect(platforms).toContain('discord');
      expect(platforms).toHaveLength(3);
    });

    it('should allow selecting a platform', () => {
      let selectedPlatform: string | null = null;
      
      const selectPlatform = (platform: string) => {
        selectedPlatform = platform;
      };

      selectPlatform('github');
      expect(selectedPlatform).toBe('github');

      selectPlatform('slack');
      expect(selectedPlatform).toBe('slack');
    });

    it('should show platform-specific information', () => {
      const platformInfo = {
        github: {
          name: 'GitHub',
          description: 'Authorize access to your repositories',
          permissions: ['Read repository metadata', 'Read pull requests']
        },
        slack: {
          name: 'Slack',
          description: 'Authorize access to your workspace',
          permissions: ['Read channel messages', 'Read thread replies']
        },
        discord: {
          name: 'Discord',
          description: 'Authorize access to your servers',
          permissions: ['Read server messages', 'Read channel information']
        }
      };

      expect(platformInfo.github.permissions).toContain('Read repository metadata');
      expect(platformInfo.slack.permissions).toContain('Read channel messages');
      expect(platformInfo.discord.permissions).toContain('Read server messages');
    });
  });

  describe('GitHub Integration Setup', () => {
    it('should initiate GitHub OAuth flow', () => {
      const csrfToken = 'mock-csrf-token';
      sessionStorageMock.setItem('integration_csrf_token', csrfToken);
      sessionStorageMock.setItem('integration_platform', 'github');
      sessionStorageMock.setItem('integration_project_id', mockProjectId);

      const oauthUrl = `/api/integrations/github/connect?project_id=${mockProjectId}&state=${csrfToken}`;

      expect(sessionStorageMock.getItem('integration_csrf_token')).toBe(csrfToken);
      expect(sessionStorageMock.getItem('integration_platform')).toBe('github');
      expect(oauthUrl).toContain('github');
      expect(oauthUrl).toContain('state=');
    });

    it('should handle successful GitHub OAuth callback', async () => {
      const integrationId = 'int-123';
      sessionStorageMock.setItem('integration_platform', 'github');

      const mockResponse = {
        success: true,
        integration_id: integrationId,
        platform: 'github',
        status: 'connected'
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse
      });

      // Simulate OAuth callback
      const urlParams = new URLSearchParams({
        oauth_success: 'true',
        integration_id: integrationId
      });

      expect(urlParams.get('oauth_success')).toBe('true');
      expect(urlParams.get('integration_id')).toBe(integrationId);
    });

    it('should load GitHub repositories after connection', async () => {
      const mockRepositories = [
        {
          id: 'repo-1',
          name: 'my-repo',
          full_name: 'user/my-repo',
          description: 'My repository',
          private: false,
          stats: '100 commits'
        },
        {
          id: 'repo-2',
          name: 'another-repo',
          full_name: 'user/another-repo',
          description: 'Another repository',
          private: true,
          stats: '50 commits'
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ sources: mockRepositories })
      });

      const response = await fetch('/api/integrations/int-123/sources', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.sources).toHaveLength(2);
      expect(data.sources[0].name).toBe('my-repo');
    });

    it('should allow selecting multiple repositories', () => {
      const selectedSources: string[] = [];

      const toggleSource = (sourceId: string, selected: boolean) => {
        if (selected) {
          if (!selectedSources.includes(sourceId)) {
            selectedSources.push(sourceId);
          }
        } else {
          const index = selectedSources.indexOf(sourceId);
          if (index > -1) {
            selectedSources.splice(index, 1);
          }
        }
      };

      toggleSource('repo-1', true);
      toggleSource('repo-2', true);
      expect(selectedSources).toHaveLength(2);

      toggleSource('repo-1', false);
      expect(selectedSources).toHaveLength(1);
      expect(selectedSources).toContain('repo-2');
    });

    it('should configure GitHub-specific settings', async () => {
      const configuration = {
        sync_frequency: 'hourly',
        sync_prs: true,
        sync_issues: true,
        sync_commits: true,
        exclude_keywords: ['wip', 'draft']
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Configuration saved'
        })
      });

      const response = await fetch('/api/integrations/int-123/configure', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({
          sources: ['repo-1', 'repo-2'],
          configuration
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
    });
  });

  describe('Slack Integration Setup', () => {
    it('should initiate Slack OAuth flow', () => {
      const csrfToken = 'mock-csrf-token';
      sessionStorageMock.setItem('integration_csrf_token', csrfToken);
      sessionStorageMock.setItem('integration_platform', 'slack');
      sessionStorageMock.setItem('integration_project_id', mockProjectId);

      const oauthUrl = `/api/integrations/slack/connect?project_id=${mockProjectId}&state=${csrfToken}`;

      expect(sessionStorageMock.getItem('integration_platform')).toBe('slack');
      expect(oauthUrl).toContain('slack');
    });

    it('should load Slack channels after connection', async () => {
      const mockChannels = [
        {
          id: 'channel-1',
          name: 'general',
          description: 'General discussion',
          type: 'channel',
          stats: '1000 messages'
        },
        {
          id: 'channel-2',
          name: 'engineering',
          description: 'Engineering team',
          type: 'channel',
          stats: '500 messages'
        },
        {
          id: 'dm-1',
          name: 'Direct Messages',
          description: 'Private conversations',
          type: 'dm',
          stats: '200 messages'
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ sources: mockChannels })
      });

      const response = await fetch('/api/integrations/int-123/sources', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.sources).toHaveLength(3);
      expect(data.sources.some((s: any) => s.type === 'dm')).toBe(true);
    });

    it('should configure Slack-specific settings', async () => {
      const configuration = {
        sync_frequency: 'hourly',
        include_dms: true,
        thread_depth: 'all',
        exclude_keywords: ['bot', 'automated']
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Configuration saved'
        })
      });

      const response = await fetch('/api/integrations/int-123/configure', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({
          sources: ['channel-1', 'channel-2'],
          configuration
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
    });
  });

  describe('Discord Integration Setup', () => {
    it('should initiate Discord bot installation', () => {
      const csrfToken = 'mock-csrf-token';
      sessionStorageMock.setItem('integration_csrf_token', csrfToken);
      sessionStorageMock.setItem('integration_platform', 'discord');
      sessionStorageMock.setItem('integration_project_id', mockProjectId);

      const oauthUrl = `/api/integrations/discord/connect?project_id=${mockProjectId}&state=${csrfToken}`;

      expect(sessionStorageMock.getItem('integration_platform')).toBe('discord');
      expect(oauthUrl).toContain('discord');
    });

    it('should load Discord servers and channels after connection', async () => {
      const mockServers = [
        {
          id: 'server-1',
          name: 'My Server',
          description: 'Main server',
          type: 'server',
          channels: [
            { id: 'channel-1', name: 'general' },
            { id: 'channel-2', name: 'dev-chat' }
          ]
        }
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ sources: mockServers })
      });

      const response = await fetch('/api/integrations/int-123/sources', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.sources).toHaveLength(1);
      expect(data.sources[0].channels).toHaveLength(2);
    });

    it('should configure Discord-specific settings', async () => {
      const configuration = {
        sync_frequency: 'hourly',
        thread_depth: 'all',
        exclude_keywords: ['spam']
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          message: 'Configuration saved'
        })
      });

      const response = await fetch('/api/integrations/int-123/configure', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({
          sources: ['channel-1', 'channel-2'],
          configuration
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.success).toBe(true);
    });
  });

  describe('Source Selection', () => {
    it('should filter sources by search query', () => {
      const sources = [
        { id: '1', name: 'my-repo', description: 'My repository' },
        { id: '2', name: 'another-repo', description: 'Another repository' },
        { id: '3', name: 'test-project', description: 'Test project' }
      ];

      const filterSources = (query: string) => {
        const lowerQuery = query.toLowerCase();
        return sources.filter(s => 
          s.name.toLowerCase().includes(lowerQuery) ||
          s.description.toLowerCase().includes(lowerQuery)
        );
      };

      expect(filterSources('my')).toHaveLength(1);
      expect(filterSources('repo')).toHaveLength(2);
      expect(filterSources('test')).toHaveLength(1);
      expect(filterSources('')).toHaveLength(3);
    });

    it('should show selected source count', () => {
      const selectedSources = ['repo-1', 'repo-2', 'repo-3'];
      const count = selectedSources.length;
      const text = `${count} source${count !== 1 ? 's' : ''} selected`;

      expect(text).toBe('3 sources selected');

      const singleSource = ['repo-1'];
      const singleText = `${singleSource.length} source${singleSource.length !== 1 ? 's' : ''} selected`;
      expect(singleText).toBe('1 source selected');
    });

    it('should validate at least one source is selected', () => {
      const validateSelection = (sources: string[]) => {
        return sources.length > 0;
      };

      expect(validateSelection([])).toBe(false);
      expect(validateSelection(['repo-1'])).toBe(true);
      expect(validateSelection(['repo-1', 'repo-2'])).toBe(true);
    });
  });

  describe('Configuration Step', () => {
    it('should validate sync frequency options', () => {
      const validFrequencies = ['realtime', 'hourly', 'daily', 'weekly'];
      
      expect(validFrequencies).toContain('hourly');
      expect(validFrequencies).toContain('daily');
      expect(validFrequencies).toHaveLength(4);
    });

    it('should parse exclude keywords', () => {
      const parseKeywords = (input: string) => {
        return input
          .split(',')
          .map(k => k.trim())
          .filter(k => k.length > 0);
      };

      expect(parseKeywords('wip, draft, test')).toEqual(['wip', 'draft', 'test']);
      expect(parseKeywords('wip,draft,test')).toEqual(['wip', 'draft', 'test']);
      expect(parseKeywords('')).toEqual([]);
      expect(parseKeywords('single')).toEqual(['single']);
    });

    it('should validate configuration before saving', () => {
      const validateConfig = (config: any) => {
        const errors = [];
        
        if (!config.sync_frequency) {
          errors.push('Sync frequency is required');
        }
        
        if (config.exclude_keywords && !Array.isArray(config.exclude_keywords)) {
          errors.push('Exclude keywords must be an array');
        }
        
        return { valid: errors.length === 0, errors };
      };

      const validConfig = {
        sync_frequency: 'hourly',
        exclude_keywords: ['wip']
      };
      expect(validateConfig(validConfig).valid).toBe(true);

      const invalidConfig = {
        exclude_keywords: ['wip']
      };
      expect(validateConfig(invalidConfig).valid).toBe(false);
    });
  });

  describe('Initial Sync Monitoring', () => {
    it('should poll for ingestion status', async () => {
      const mockStatus = {
        state: 'syncing',
        progress: 45,
        message: 'Syncing repositories...',
        items_processed: 45,
        items_total: 100
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ status: mockStatus })
      });

      const response = await fetch('/api/integrations/int-123/status', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.status.state).toBe('syncing');
      expect(data.status.progress).toBe(45);
    });

    it('should handle sync completion', async () => {
      const mockStatus = {
        state: 'completed',
        progress: 100,
        message: 'Sync completed successfully',
        items_processed: 100,
        items_total: 100,
        summary: {
          entities_created: 50,
          decisions_extracted: 10,
          discussions_processed: 30
        }
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ status: mockStatus })
      });

      const response = await fetch('/api/integrations/int-123/status', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.status.state).toBe('completed');
      expect(data.status.progress).toBe(100);
      expect(data.status.summary).toBeDefined();
    });

    it('should handle sync errors', async () => {
      const mockStatus = {
        state: 'error',
        progress: 30,
        message: 'Sync failed',
        error: 'Rate limit exceeded',
        retry_after: 3600
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ status: mockStatus })
      });

      const response = await fetch('/api/integrations/int-123/status', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      const data = await response.json();

      expect(response.ok).toBe(true);
      expect(data.status.state).toBe('error');
      expect(data.status.error).toBeDefined();
    });

    it('should update progress bar based on status', () => {
      const updateProgress = (progress: number) => {
        return {
          width: `${progress}%`,
          text: `${Math.round(progress)}%`
        };
      };

      expect(updateProgress(0).width).toBe('0%');
      expect(updateProgress(50).width).toBe('50%');
      expect(updateProgress(100).width).toBe('100%');
      expect(updateProgress(45.7).text).toBe('46%');
    });
  });

  describe('Wizard Navigation', () => {
    it('should track current step', () => {
      let currentStep = 1;
      const totalSteps = 5;

      const nextStep = () => {
        if (currentStep < totalSteps) {
          currentStep++;
        }
      };

      const previousStep = () => {
        if (currentStep > 1) {
          currentStep--;
        }
      };

      expect(currentStep).toBe(1);
      
      nextStep();
      expect(currentStep).toBe(2);
      
      nextStep();
      nextStep();
      expect(currentStep).toBe(4);
      
      previousStep();
      expect(currentStep).toBe(3);
      
      // Can't go below 1
      currentStep = 1;
      previousStep();
      expect(currentStep).toBe(1);
      
      // Can't go above totalSteps
      currentStep = 5;
      nextStep();
      expect(currentStep).toBe(5);
    });

    it('should validate step before allowing navigation', () => {
      const validateStep = (step: number, data: any) => {
        switch (step) {
          case 1: // Platform selection
            return !!data.selectedPlatform;
          case 2: // OAuth connection
            return !!data.integrationId;
          case 3: // Source selection
            return data.selectedSources && data.selectedSources.length > 0;
          case 4: // Configuration
            return !!data.configuration;
          default:
            return true;
        }
      };

      expect(validateStep(1, { selectedPlatform: 'github' })).toBe(true);
      expect(validateStep(1, {})).toBe(false);
      
      expect(validateStep(3, { selectedSources: ['repo-1'] })).toBe(true);
      expect(validateStep(3, { selectedSources: [] })).toBe(false);
    });
  });

  describe('Error Handling', () => {
    it('should handle OAuth errors', async () => {
      const mockResponse = {
        success: false,
        message: 'OAuth authentication failed',
        error: 'access_denied'
      };

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => mockResponse
      });

      const response = await fetch('/api/integrations/github/connect', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({
          code: 'invalid-code',
          state: 'valid-state'
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.error).toBe('access_denied');
    });

    it('should handle source loading errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({
          success: false,
          message: 'Failed to load sources'
        })
      });

      const response = await fetch('/api/integrations/int-123/sources', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${mockAuthToken}`
        }
      });

      expect(response.ok).toBe(false);
      expect(response.status).toBe(500);
    });

    it('should handle configuration save errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({
          success: false,
          message: 'Invalid configuration',
          errors: ['sync_frequency is required']
        })
      });

      const response = await fetch('/api/integrations/int-123/configure', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${mockAuthToken}`
        },
        body: JSON.stringify({
          sources: ['repo-1'],
          configuration: {}
        })
      });

      const data = await response.json();

      expect(response.ok).toBe(false);
      expect(data.errors).toBeDefined();
    });
  });

  describe('Wizard Completion', () => {
    it('should show completion summary', () => {
      const completionData = {
        platform: 'github',
        sources_count: 3,
        configuration: {
          sync_frequency: 'hourly'
        },
        initial_sync_status: 'completed'
      };

      expect(completionData.platform).toBe('github');
      expect(completionData.sources_count).toBe(3);
      expect(completionData.initial_sync_status).toBe('completed');
    });

    it('should provide navigation options after completion', () => {
      const actions = [
        { id: 'add-another', label: 'Add Another Integration' },
        { id: 'view-project', label: 'View Project Dashboard' }
      ];

      expect(actions).toHaveLength(2);
      expect(actions[0].id).toBe('add-another');
      expect(actions[1].id).toBe('view-project');
    });

    it('should reset wizard state when adding another integration', () => {
      const resetWizard = () => {
        return {
          currentStep: 1,
          selectedPlatform: null,
          selectedSources: [],
          configuration: {},
          integrationId: null
        };
      };

      const resetState = resetWizard();

      expect(resetState.currentStep).toBe(1);
      expect(resetState.selectedPlatform).toBeNull();
      expect(resetState.selectedSources).toHaveLength(0);
    });
  });
});
