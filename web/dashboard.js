// Dashboard Module
// Handles project management, integration status, and metrics visualization

class DashboardManager {
    constructor() {
        this.authManager = new AuthManager();
        this.apiBaseUrl = '/api';
        this.currentProject = null;
        this.refreshInterval = null;
    }

    // Check authentication and redirect if needed
    checkAuth() {
        if (!this.authManager.isAuthenticated()) {
            window.location.href = '/login';
            return false;
        }
        return true;
    }

    // Get auth headers for API calls
    getAuthHeaders() {
        const token = this.authManager.getToken();
        return {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        };
    }

    // Fetch all projects for the current user
    async fetchProjects() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                if (response.status === 401) {
                    this.authManager.clearSession();
                    window.location.href = '/login';
                    return null;
                }
                throw new Error('Failed to fetch projects');
            }

            const data = await response.json();
            return data.projects || [];
        } catch (error) {
            console.error('Error fetching projects:', error);
            return null;
        }
    }

    // Create a new project
    async createProject(projectData) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects`, {
                method: 'POST',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(projectData)
            });


            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to create project');
            }

            const data = await response.json();
            return { success: true, project: data.project };
        } catch (error) {
            console.error('Error creating project:', error);
            return { success: false, error: error.message };
        }
    }

    // Fetch project details
    async fetchProjectDetails(projectId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects/${projectId}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                throw new Error('Failed to fetch project details');
            }

            const data = await response.json();
            return data.project;
        } catch (error) {
            console.error('Error fetching project details:', error);
            return null;
        }
    }

    // Fetch project integrations
    async fetchProjectIntegrations(projectId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects/${projectId}/integrations`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                throw new Error('Failed to fetch integrations');
            }

            const data = await response.json();
            return data.integrations || [];
        } catch (error) {
            console.error('Error fetching integrations:', error);
            return [];
        }
    }

    // Fetch project metrics
    async fetchProjectMetrics(projectId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects/${projectId}/metrics`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                throw new Error('Failed to fetch metrics');
            }

            const data = await response.json();
            return data.metrics;
        } catch (error) {
            console.error('Error fetching metrics:', error);
            return null;
        }
    }

    // Fetch project statistics
    async fetchProjectStats(projectId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects/${projectId}/stats`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                throw new Error('Failed to fetch stats');
            }

            const data = await response.json();
            return data.stats;
        } catch (error) {
            console.error('Error fetching stats:', error);
            return null;
        }
    }

    // Fetch recent activity
    async fetchRecentActivity(projectId, limit = 10) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects/${projectId}/activity?limit=${limit}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                throw new Error('Failed to fetch activity');
            }

            const data = await response.json();
            return data.activities || [];
        } catch (error) {
            console.error('Error fetching activity:', error);
            return [];
        }
    }

    // Connect integration
    async connectIntegration(projectId, platform) {
        try {
            // Generate CSRF token
            const csrfToken = this.authManager.generateCSRFToken();
            sessionStorage.setItem('integration_csrf_token', csrfToken);
            sessionStorage.setItem('integration_project_id', projectId);
            
            // Redirect to integration OAuth flow
            window.location.href = `${this.apiBaseUrl}/integrations/${platform}/connect?project_id=${projectId}&state=${csrfToken}`;
        } catch (error) {
            console.error('Error connecting integration:', error);
            return { success: false, error: error.message };
        }
    }

    // Delete project
    async deleteProject(projectId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/projects/${projectId}`, {
                method: 'DELETE',
                headers: this.getAuthHeaders()
            });

            if (!response.ok) {
                throw new Error('Failed to delete project');
            }

            return { success: true };
        } catch (error) {
            console.error('Error deleting project:', error);
            return { success: false, error: error.message };
        }
    }

    // Format date for display
    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now - date;
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMs / 3600000);
        const diffDays = Math.floor(diffMs / 86400000);

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
        if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
        if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
        
        return date.toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric', 
            year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined 
        });
    }

    // Format number with commas
    formatNumber(num) {
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    }

    // Start auto-refresh for real-time updates
    startAutoRefresh(callback, interval = 30000) {
        this.stopAutoRefresh();
        this.refreshInterval = setInterval(callback, interval);
    }

    // Stop auto-refresh
    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }
}

// UI Helper Functions for Dashboard
class DashboardUI {
    static renderProjectCard(project) {
        const integrations = project.integrations || [];
        const stats = project.stats || {};
        
        const integrationsHTML = integrations.length > 0 
            ? integrations.map(int => `
                <div class="integration-badge ${int.platform}">
                    ${DashboardUI.getIntegrationIcon(int.platform)}
                    <span>${int.platform}</span>
                </div>
            `).join('')
            : '<span style="color: var(--text-tertiary); font-size: 13px;">No integrations</span>';

        return `
            <a href="/project?id=${project.id}" class="project-card">
                <div class="project-card-header">
                    <div>
                        <h3 class="project-card-title">${DashboardUI.escapeHtml(project.name)}</h3>
                        <p class="project-card-description">${DashboardUI.escapeHtml(project.description || 'No description')}</p>
                    </div>
                    <button class="project-card-menu" onclick="event.preventDefault(); event.stopPropagation(); DashboardUI.showProjectMenu(${project.id})">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="1"></circle>
                            <circle cx="12" cy="5" r="1"></circle>
                            <circle cx="12" cy="19" r="1"></circle>
                        </svg>
                    </button>
                </div>
                <div class="project-card-integrations">
                    ${integrationsHTML}
                </div>
                <div class="project-card-stats">
                    <div class="project-stat">
                        <span class="project-stat-value">${stats.entities || 0}</span>
                        <span class="project-stat-label">Entities</span>
                    </div>
                    <div class="project-stat">
                        <span class="project-stat-value">${stats.decisions || 0}</span>
                        <span class="project-stat-label">Decisions</span>
                    </div>
                    <div class="project-stat">
                        <span class="project-stat-value">${stats.discussions || 0}</span>
                        <span class="project-stat-label">Discussions</span>
                    </div>
                </div>
            </a>
        `;
    }

    static getIntegrationIcon(platform) {
        const icons = {
            github: '<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>',
            slack: '<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52z"/></svg>',
            discord: '<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028c.462-.63.874-1.295 1.226-1.994a.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z"/></svg>'
        };
        return icons[platform] || '';
    }

    static showProjectMenu(projectId) {
        // TODO: Implement project menu (edit, delete, settings)
        console.log('Show menu for project:', projectId);
    }

    static escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    static renderActivityItem(activity) {
        return `
            <div class="activity-item">
                <div class="activity-icon">
                    ${DashboardUI.getActivityIcon(activity.type)}
                </div>
                <div class="activity-content">
                    <h4 class="activity-title">${DashboardUI.escapeHtml(activity.title)}</h4>
                    <p class="activity-description">${DashboardUI.escapeHtml(activity.description)}</p>
                    <span class="activity-time">${activity.time}</span>
                </div>
            </div>
        `;
    }

    static getActivityIcon(type) {
        const icons = {
            sync: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.2"/></svg>',
            decision: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>',
            discussion: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>',
            integration: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path></svg>'
        };
        return icons[type] || icons.sync;
    }
}

// Export for use in HTML pages
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { DashboardManager, DashboardUI };
}
