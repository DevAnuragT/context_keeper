// Integration Wizard Module
// Handles multi-step integration setup flow

class IntegrationWizard {
    constructor() {
        this.authManager = new AuthManager();
        this.dashboardManager = new DashboardManager();
        this.apiBaseUrl = '/api';
        this.currentStep = 1;
        this.totalSteps = 5;
        this.selectedPlatform = null;
        this.selectedSources = [];
        this.configuration = {};
        this.projectId = null;
        this.integrationId = null;
        this.pollInterval = null;
    }

    // Initialize wizard
    init() {
        // Check authentication
        if (!this.authManager.isAuthenticated()) {
            window.location.href = '/login';
            return;
        }

        // Get project ID from URL
        const urlParams = new URLSearchParams(window.location.search);
        this.projectId = urlParams.get('project_id');
        
        if (!this.projectId) {
            window.location.href = '/dashboard';
            return;
        }

        // Load project info
        this.loadProjectInfo();

        // Setup event listeners
        this.setupEventListeners();

        // Initialize user menu
        this.initializeUserMenu();
    }

    // Load project information
    async loadProjectInfo() {
        const project = await this.dashboardManager.fetchProjectDetails(this.projectId);
        if (project) {
            document.getElementById('project-name-breadcrumb').textContent = project.name;
        }
    }

    // Setup all event listeners
    setupEventListeners() {
        // Platform selection
        document.querySelectorAll('.platform-card').forEach(card => {
            card.addEventListener('click', () => {
                this.selectPlatform(card.dataset.platform);
            });
        });

        // Navigation buttons
        document.getElementById('back-btn').addEventListener('click', () => this.previousStep());
        document.getElementById('next-btn').addEventListener('click', () => this.nextStep());

        // OAuth connection
        document.getElementById('connect-oauth-btn').addEventListener('click', () => {
            this.initiateOAuthFlow();
        });

        // Source search
        document.getElementById('source-search-input').addEventListener('input', (e) => {
            this.filterSources(e.target.value);
        });

        // Completion actions
        document.getElementById('add-another-btn').addEventListener('click', () => {
            this.resetWizard();
        });

        document.getElementById('view-project-btn').addEventListener('click', () => {
            window.location.href = `/project?id=${this.projectId}`;
        });

        // User menu
        document.getElementById('user-menu-btn').addEventListener('click', (e) => {
            e.stopPropagation();
            document.getElementById('user-menu-dropdown').classList.toggle('active');
        });

        document.addEventListener('click', () => {
            document.getElementById('user-menu-dropdown').classList.remove('active');
        });

        document.getElementById('logout-btn').addEventListener('click', async () => {
            await this.authManager.logout();
            window.location.href = '/login';
        });

        // Handle OAuth callback
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('oauth_success') === 'true') {
            this.handleOAuthCallback();
        }
    }

    // Initialize user menu with user data
    initializeUserMenu() {
        const userData = this.authManager.getUserData();
        if (userData) {
            document.getElementById('user-name').textContent = userData.name || userData.email;
            document.getElementById('menu-user-name').textContent = userData.name || userData.email;
            document.getElementById('menu-user-email').textContent = userData.email;
            
            const initials = (userData.name || userData.email)
                .split(' ')
                .map(n => n[0])
                .join('')
                .toUpperCase()
                .substring(0, 2);
            document.getElementById('user-initials').textContent = initials;
        }
    }

    // Select platform
    selectPlatform(platform) {
        this.selectedPlatform = platform;
        
        // Update UI
        document.querySelectorAll('.platform-card').forEach(card => {
            card.classList.remove('selected');
        });
        document.querySelector(`[data-platform="${platform}"]`).classList.add('selected');

        // Enable next button
        document.getElementById('next-btn').disabled = false;
    }

    // Navigate to next step
    async nextStep() {
        if (this.currentStep === 1) {
            // Moving from platform selection to OAuth
            this.setupOAuthStep();
        } else if (this.currentStep === 2) {
            // Moving from OAuth to source selection
            await this.loadSources();
        } else if (this.currentStep === 3) {
            // Moving from source selection to configuration
            this.setupConfigurationStep();
        } else if (this.currentStep === 4) {
            // Moving from configuration to completion
            await this.saveIntegration();
        }

        if (this.currentStep < this.totalSteps) {
            this.currentStep++;
            this.updateStepUI();
        }
    }

    // Navigate to previous step
    previousStep() {
        if (this.currentStep > 1) {
            this.currentStep--;
            this.updateStepUI();
        }
    }

    // Update step UI
    updateStepUI() {
        // Update step indicators
        document.querySelectorAll('.wizard-step').forEach((step, index) => {
            const stepNum = index + 1;
            if (stepNum < this.currentStep) {
                step.classList.add('completed');
                step.classList.remove('active');
            } else if (stepNum === this.currentStep) {
                step.classList.add('active');
                step.classList.remove('completed');
            } else {
                step.classList.remove('active', 'completed');
            }
        });

        // Update panels
        document.querySelectorAll('.wizard-panel').forEach((panel, index) => {
            if (index + 1 === this.currentStep) {
                panel.classList.add('active');
            } else {
                panel.classList.remove('active');
            }
        });

        // Update navigation buttons
        const backBtn = document.getElementById('back-btn');
        const nextBtn = document.getElementById('next-btn');

        if (this.currentStep === 1) {
            backBtn.style.display = 'none';
        } else {
            backBtn.style.display = 'flex';
        }

        if (this.currentStep === this.totalSteps) {
            nextBtn.style.display = 'none';
        } else {
            nextBtn.style.display = 'flex';
        }

        // Disable next button by default (will be enabled when step is valid)
        if (this.currentStep !== 1 || !this.selectedPlatform) {
            nextBtn.disabled = true;
        }
    }

    // Setup OAuth connection step
    setupOAuthStep() {
        const platformConfig = this.getPlatformConfig(this.selectedPlatform);
        
        // Update UI
        document.getElementById('connection-platform-icon').innerHTML = platformConfig.icon;
        document.getElementById('connection-title').textContent = `Connect to ${platformConfig.name}`;
        document.getElementById('connection-description').textContent = platformConfig.description;
        
        // Update permissions list
        const permissionsList = document.getElementById('permissions-list');
        permissionsList.innerHTML = platformConfig.permissions.map(perm => `
            <li>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                ${perm}
            </li>
        `).join('');

        document.getElementById('connect-btn-text').textContent = `Connect ${platformConfig.name}`;
    }

    // Get platform configuration
    getPlatformConfig(platform) {
        const configs = {
            github: {
                name: 'GitHub',
                icon: '<svg width="48" height="48" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>',
                description: 'Authorize access to your repositories',
                permissions: [
                    'Read repository metadata',
                    'Read pull requests and issues',
                    'Read commit history',
                    'Read file contents'
                ]
            },
            slack: {
                name: 'Slack',
                icon: '<svg width="48" height="48" viewBox="0 0 24 24" fill="currentColor"><path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52z"/></svg>',
                description: 'Authorize access to your workspace',
                permissions: [
                    'Read channel messages',
                    'Read thread replies',
                    'Read user information',
                    'Access workspace metadata'
                ]
            },
            discord: {
                name: 'Discord',
                icon: '<svg width="48" height="48" viewBox="0 0 24 24" fill="currentColor"><path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z"/></svg>',
                description: 'Authorize access to your servers',
                permissions: [
                    'Read server messages',
                    'Read channel information',
                    'Read user information',
                    'Access server metadata'
                ]
            }
        };
        return configs[platform];
    }

    // Initiate OAuth flow
    initiateOAuthFlow() {
        const btn = document.getElementById('connect-oauth-btn');
        btn.disabled = true;
        btn.innerHTML = '<span class="spinner"></span> Connecting...';

        // Update status
        const statusEl = document.getElementById('connection-status');
        statusEl.innerHTML = `
            <div class="status-indicator connecting">
                <div class="spinner"></div>
                <span>Opening authorization window...</span>
            </div>
        `;

        // Generate CSRF token
        const csrfToken = this.authManager.generateCSRFToken();
        sessionStorage.setItem('integration_csrf_token', csrfToken);
        sessionStorage.setItem('integration_platform', this.selectedPlatform);
        sessionStorage.setItem('integration_project_id', this.projectId);

        // Redirect to OAuth flow
        const redirectUrl = `${this.apiBaseUrl}/integrations/${this.selectedPlatform}/connect?project_id=${this.projectId}&state=${csrfToken}`;
        window.location.href = redirectUrl;
    }

    // Handle OAuth callback
    async handleOAuthCallback() {
        const urlParams = new URLSearchParams(window.location.search);
        const platform = sessionStorage.getItem('integration_platform');
        const integrationId = urlParams.get('integration_id');

        if (!platform || !integrationId) {
            this.showOAuthError('Invalid OAuth callback');
            return;
        }

        this.selectedPlatform = platform;
        this.integrationId = integrationId;

        // Update status
        const statusEl = document.getElementById('connection-status');
        statusEl.innerHTML = `
            <div class="status-indicator connected">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"></polyline>
                </svg>
                <span>Successfully connected!</span>
            </div>
        `;

        // Enable next button
        document.getElementById('next-btn').disabled = false;

        // Clean up session storage
        sessionStorage.removeItem('integration_csrf_token');
        sessionStorage.removeItem('integration_platform');
        sessionStorage.removeItem('integration_project_id');

        // Auto-advance after a short delay
        setTimeout(() => {
            this.nextStep();
        }, 1500);
    }

    // Show OAuth error
    showOAuthError(message) {
        const statusEl = document.getElementById('connection-status');
        statusEl.innerHTML = `
            <div class="status-indicator error">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="15" y1="9" x2="9" y2="15"></line>
                    <line x1="9" y1="9" x2="15" y2="15"></line>
                </svg>
                <span>${message}</span>
            </div>
        `;

        const btn = document.getElementById('connect-oauth-btn');
        btn.disabled = false;
        btn.innerHTML = `
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
            </svg>
            Try Again
        `;
    }

    // Load sources from platform
    async loadSources() {
        const sourceList = document.getElementById('source-list');
        sourceList.innerHTML = `
            <div class="loading-sources">
                <div class="spinner"></div>
                <p>Loading sources...</p>
            </div>
        `;

        try {
            const response = await fetch(
                `${this.apiBaseUrl}/integrations/${this.integrationId}/sources`,
                {
                    method: 'GET',
                    headers: this.getAuthHeaders()
                }
            );

            if (!response.ok) {
                throw new Error('Failed to load sources');
            }

            const data = await response.json();
            this.renderSources(data.sources || []);
        } catch (error) {
            console.error('Error loading sources:', error);
            sourceList.innerHTML = `
                <div class="error-state">
                    <p>Failed to load sources. Please try again.</p>
                    <button class="btn-secondary" onclick="wizard.loadSources()">Retry</button>
                </div>
            `;
        }
    }

    // Render sources list
    renderSources(sources) {
        const sourceList = document.getElementById('source-list');
        
        if (sources.length === 0) {
            sourceList.innerHTML = `
                <div class="empty-state">
                    <p>No sources available</p>
                </div>
            `;
            return;
        }

        sourceList.innerHTML = sources.map(source => `
            <label class="source-item" data-source-id="${source.id}">
                <input type="checkbox" class="source-checkbox" value="${source.id}">
                <div class="source-info">
                    <div class="source-icon">
                        ${this.getSourceIcon(source.type)}
                    </div>
                    <div class="source-details">
                        <div class="source-name">${this.escapeHtml(source.name)}</div>
                        <div class="source-meta">${source.description || ''}</div>
                    </div>
                </div>
                <div class="source-stats">
                    ${source.stats ? `<span>${source.stats}</span>` : ''}
                </div>
            </label>
        `).join('');

        // Add event listeners
        document.querySelectorAll('.source-checkbox').forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                this.toggleSource(e.target.value, e.target.checked);
            });
        });

        // Update description based on platform
        const descriptions = {
            github: 'Select repositories to sync',
            slack: 'Select channels to sync',
            discord: 'Select servers and channels to sync'
        };
        document.getElementById('source-description').textContent = 
            descriptions[this.selectedPlatform] || 'Choose sources to sync';
    }

    // Get source icon based on type
    getSourceIcon(type) {
        const icons = {
            repository: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>',
            channel: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>',
            server: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect><rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect></svg>'
        };
        return icons[type] || icons.repository;
    }

    // Toggle source selection
    toggleSource(sourceId, selected) {
        if (selected) {
            if (!this.selectedSources.includes(sourceId)) {
                this.selectedSources.push(sourceId);
            }
        } else {
            this.selectedSources = this.selectedSources.filter(id => id !== sourceId);
        }

        // Update summary
        const count = this.selectedSources.length;
        document.getElementById('selected-count').textContent = 
            `${count} source${count !== 1 ? 's' : ''} selected`;

        // Enable/disable next button
        document.getElementById('next-btn').disabled = count === 0;
    }

    // Filter sources
    filterSources(query) {
        const lowerQuery = query.toLowerCase();
        document.querySelectorAll('.source-item').forEach(item => {
            const name = item.querySelector('.source-name').textContent.toLowerCase();
            const meta = item.querySelector('.source-meta').textContent.toLowerCase();
            
            if (name.includes(lowerQuery) || meta.includes(lowerQuery)) {
                item.style.display = 'flex';
            } else {
                item.style.display = 'none';
            }
        });
    }

    // Setup configuration step
    setupConfigurationStep() {
        // Show platform-specific config
        document.querySelectorAll('[id$="-config"]').forEach(el => {
            el.style.display = 'none';
        });
        
        const configEl = document.getElementById(`${this.selectedPlatform}-config`);
        if (configEl) {
            configEl.style.display = 'block';
        }

        // Enable next button
        document.getElementById('next-btn').disabled = false;
    }

    // Save integration configuration
    async saveIntegration() {
        // Collect configuration
        this.configuration = {
            sync_frequency: document.querySelector('input[name="sync-frequency"]:checked').value,
            exclude_keywords: document.querySelector('input[name="exclude-keywords"]').value
                .split(',')
                .map(k => k.trim())
                .filter(k => k.length > 0)
        };

        // Platform-specific config
        if (this.selectedPlatform === 'github') {
            this.configuration.sync_prs = document.querySelector('input[name="sync-prs"]').checked;
            this.configuration.sync_issues = document.querySelector('input[name="sync-issues"]').checked;
            this.configuration.sync_commits = document.querySelector('input[name="sync-commits"]').checked;
        } else if (this.selectedPlatform === 'slack') {
            this.configuration.include_dms = document.querySelector('input[name="include-dms"]').checked;
            this.configuration.thread_depth = document.querySelector('select[name="thread-depth"]').value;
        } else if (this.selectedPlatform === 'discord') {
            this.configuration.thread_depth = document.querySelector('select[name="thread-depth"]').value;
        }

        try {
            const response = await fetch(
                `${this.apiBaseUrl}/integrations/${this.integrationId}/configure`,
                {
                    method: 'POST',
                    headers: this.getAuthHeaders(),
                    body: JSON.stringify({
                        sources: this.selectedSources,
                        configuration: this.configuration
                    })
                }
            );

            if (!response.ok) {
                throw new Error('Failed to save configuration');
            }

            // Update completion UI
            document.getElementById('sources-selected-text').textContent = 
                `${this.selectedSources.length} source${this.selectedSources.length !== 1 ? 's' : ''} selected`;

            // Start polling for ingestion status
            this.startIngestionPolling();
        } catch (error) {
            console.error('Error saving integration:', error);
            alert('Failed to save integration. Please try again.');
        }
    }

    // Start polling for ingestion status
    startIngestionPolling() {
        this.pollInterval = setInterval(async () => {
            await this.updateIngestionStatus();
        }, 2000);

        // Initial update
        this.updateIngestionStatus();
    }

    // Update ingestion status
    async updateIngestionStatus() {
        try {
            const response = await fetch(
                `${this.apiBaseUrl}/integrations/${this.integrationId}/status`,
                {
                    method: 'GET',
                    headers: this.getAuthHeaders()
                }
            );

            if (!response.ok) {
                throw new Error('Failed to fetch status');
            }

            const data = await response.json();
            const status = data.status;

            // Update progress bar
            const progress = status.progress || 0;
            document.getElementById('ingestion-progress').style.width = `${progress}%`;
            document.getElementById('ingestion-progress-text').textContent = `${Math.round(progress)}%`;

            // Update message
            document.getElementById('ingestion-message').textContent = 
                status.message || 'Processing...';

            // Check if complete
            if (status.state === 'completed') {
                this.handleIngestionComplete();
            } else if (status.state === 'error') {
                this.handleIngestionError(status.error);
            }
        } catch (error) {
            console.error('Error fetching ingestion status:', error);
        }
    }

    // Handle ingestion completion
    handleIngestionComplete() {
        // Stop polling
        if (this.pollInterval) {
            clearInterval(this.pollInterval);
            this.pollInterval = null;
        }

        // Update UI
        const syncStatusItem = document.getElementById('sync-status-item');
        syncStatusItem.classList.add('completed');
        syncStatusItem.innerHTML = `
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
            <span>Initial sync complete</span>
        `;

        document.getElementById('ingestion-message').textContent = 
            'All data has been successfully synced!';
    }

    // Handle ingestion error
    handleIngestionError(error) {
        // Stop polling
        if (this.pollInterval) {
            clearInterval(this.pollInterval);
            this.pollInterval = null;
        }

        // Update UI
        const completionIcon = document.getElementById('completion-icon');
        completionIcon.classList.remove('success');
        completionIcon.classList.add('error');
        completionIcon.innerHTML = `
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="15" y1="9" x2="9" y2="15"></line>
                <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
        `;

        document.getElementById('completion-title').textContent = 'Sync Error';
        document.getElementById('completion-description').textContent = 
            'There was an error during the initial sync';

        document.getElementById('ingestion-message').textContent = error || 'Unknown error occurred';
    }

    // Reset wizard
    resetWizard() {
        this.currentStep = 1;
        this.selectedPlatform = null;
        this.selectedSources = [];
        this.configuration = {};
        this.integrationId = null;

        // Clear selections
        document.querySelectorAll('.platform-card').forEach(card => {
            card.classList.remove('selected');
        });

        // Reset UI
        this.updateStepUI();
        document.getElementById('next-btn').disabled = true;

        // Clear search
        document.getElementById('source-search-input').value = '';
    }

    // Get auth headers
    getAuthHeaders() {
        const token = this.authManager.getToken();
        return {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        };
    }

    // Escape HTML
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize wizard on page load
let wizard;
document.addEventListener('DOMContentLoaded', () => {
    wizard = new IntegrationWizard();
    wizard.init();
});
