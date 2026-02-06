package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/models"
)

// PlatformConnectorManager defines the interface for managing platform connectors
type PlatformConnectorManager interface {
	GetConnector(platform string) (PlatformConnector, error)
}

// PlatformConnector defines the interface for platform connectors (avoiding import cycle)
type PlatformConnector interface {
	FetchEvents(ctx context.Context, since time.Time, limit int) ([]PlatformEvent, error)
	NormalizeData(ctx context.Context, events []PlatformEvent) ([]PlatformNormalizedEvent, error)
}

// ContextProcessorService handles context processing operations
type ContextProcessorService interface {
	ProcessEvents(ctx context.Context, events []NormalizedEvent) (*ProcessingResult, error)
}

// PlatformEvent represents a raw event from a platform
type PlatformEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Timestamp   time.Time             `json:"timestamp"`
	Author      string                `json:"author"`
	Content     string                `json:"content"`
	Title       string                `json:"title,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	References  []string              `json:"references"`
	Platform    string                `json:"platform"`
}

// PlatformNormalizedEvent represents a normalized platform event
type PlatformNormalizedEvent struct {
	PlatformID   string                 `json:"platform_id"`
	EventType    string                 `json:"event_type"`
	Timestamp    time.Time             `json:"timestamp"`
	Author       string                `json:"author"`
	Content      string                `json:"content"`
	Title        string                `json:"title,omitempty"`
	ThreadID     *string               `json:"thread_id,omitempty"`
	ParentID     *string               `json:"parent_id,omitempty"`
	FileRefs     []string              `json:"file_refs"`
	FeatureRefs  []string              `json:"feature_refs"`
	Labels       []string              `json:"labels,omitempty"`
	State        string                `json:"state,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	Platform     string                `json:"platform"`
}

// IngestionTask represents an active ingestion task
type IngestionTask struct {
	IntegrationID string
	ProjectID     string
	Platform      string
	Status        IngestionTaskStatus
	StartedAt     time.Time
	LastSyncAt    *time.Time
	ErrorCount    int
	LastError     *string
	Cancel        context.CancelFunc
}

// IngestionTaskStatus represents the status of an ingestion task
type IngestionTaskStatus string

const (
	TaskStatusRunning   IngestionTaskStatus = "running"
	TaskStatusPaused    IngestionTaskStatus = "paused"
	TaskStatusFailed    IngestionTaskStatus = "failed"
	TaskStatusCompleted IngestionTaskStatus = "completed"
)

// IngestionOrchestratorImpl implements IngestionOrchestrator
type IngestionOrchestratorImpl struct {
	store              RepositoryStore
	connectorManager   PlatformConnectorManager
	contextProcessor   ContextProcessorService
	encryptSvc         EncryptionService
	logger             Logger
	
	// Orchestration state
	mu                 sync.RWMutex
	activeIngestions   map[string]*IngestionTask // integrationID -> task
	orchestratorCtx    context.Context
	orchestratorCancel context.CancelFunc
	orchestratorWg     sync.WaitGroup
	
	// Configuration
	maxRetries         int
	retryBackoff       time.Duration
	healthCheckInterval time.Duration
	deduplicationWindow time.Duration
}

// NewIngestionOrchestrator creates a new ingestion orchestrator
func NewIngestionOrchestrator(
	store RepositoryStore,
	connectorManager PlatformConnectorManager,
	contextProcessor ContextProcessorService,
	encryptSvc EncryptionService,
	logger Logger,
) IngestionOrchestrator {
	return &IngestionOrchestratorImpl{
		store:               store,
		connectorManager:    connectorManager,
		contextProcessor:    contextProcessor,
		encryptSvc:          encryptSvc,
		logger:              logger,
		activeIngestions:    make(map[string]*IngestionTask),
		maxRetries:          3,
		retryBackoff:        time.Minute * 5,
		healthCheckInterval: time.Minute * 5,
		deduplicationWindow: time.Hour * 24,
	}
}

// StartProjectIngestion starts ingestion for all integrations in a project
func (io *IngestionOrchestratorImpl) StartProjectIngestion(ctx context.Context, projectID string) error {
	// Get all active integrations for the project
	integrations, err := io.store.GetProjectIntegrations(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project integrations: %w", err)
	}

	var errors []error
	for _, integration := range integrations {
		if integration.Status == string(models.IntegrationStatusActive) {
			if err := io.StartIntegrationIngestion(ctx, integration.ID); err != nil {
				errors = append(errors, fmt.Errorf("failed to start integration %s: %w", integration.ID, err))
				io.logger.Error("Failed to start integration ingestion", err, map[string]interface{}{
					"project_id":     projectID,
					"integration_id": integration.ID,
					"platform":       integration.Platform,
				})
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start some integrations: %v", errors)
	}

	io.logger.Info("Started project ingestion", map[string]interface{}{
		"project_id":        projectID,
		"integration_count": len(integrations),
	})

	return nil
}

// StartIntegrationIngestion starts ingestion for a specific integration
func (io *IngestionOrchestratorImpl) StartIntegrationIngestion(ctx context.Context, integrationID string) error {
	// Check if already running
	io.mu.RLock()
	if task, exists := io.activeIngestions[integrationID]; exists {
		io.mu.RUnlock()
		return fmt.Errorf("ingestion already running for integration %s (status: %s)", integrationID, task.Status)
	}
	io.mu.RUnlock()

	// Get integration details
	integration, err := io.store.GetProjectIntegration(ctx, integrationID)
	if err != nil {
		return fmt.Errorf("failed to get integration: %w", err)
	}

	if integration.Status != string(models.IntegrationStatusActive) {
		return fmt.Errorf("integration is not active (status: %s)", integration.Status)
	}

	// Create task context
	taskCtx, taskCancel := context.WithCancel(context.Background())

	// Create ingestion task
	task := &IngestionTask{
		IntegrationID: integrationID,
		ProjectID:     integration.ProjectID,
		Platform:      integration.Platform,
		Status:        TaskStatusRunning,
		StartedAt:     time.Now(),
		ErrorCount:    0,
		Cancel:        taskCancel,
	}

	// Register task
	io.mu.Lock()
	io.activeIngestions[integrationID] = task
	io.mu.Unlock()

	// Start ingestion in background
	go io.runIntegrationIngestion(taskCtx, task, integration)

	io.logger.Info("Started integration ingestion", map[string]interface{}{
		"integration_id": integrationID,
		"project_id":     integration.ProjectID,
		"platform":       integration.Platform,
	})

	return nil
}

// runIntegrationIngestion runs the ingestion process for an integration
func (io *IngestionOrchestratorImpl) runIntegrationIngestion(ctx context.Context, task *IngestionTask, integration *models.ProjectIntegration) {
	defer func() {
		// Clean up task on completion
		io.mu.Lock()
		delete(io.activeIngestions, task.IntegrationID)
		io.mu.Unlock()
	}()

	// Get connector for platform
	connector, err := io.connectorManager.GetConnector(integration.Platform)
	if err != nil {
		io.handleIngestionError(ctx, task, integration, fmt.Errorf("failed to get connector: %w", err))
		return
	}

	// Get data sources for integration
	dataSources, err := io.store.GetProjectDataSourcesByIntegration(ctx, integration.ID)
	if err != nil {
		io.handleIngestionError(ctx, task, integration, fmt.Errorf("failed to get data sources: %w", err))
		return
	}

	if len(dataSources) == 0 {
		io.logger.Info("No data sources configured for integration", map[string]interface{}{
			"integration_id": integration.ID,
		})
		task.Status = TaskStatusCompleted
		return
	}

	// Get sync checkpoint
	checkpoint := integration.SyncCheckpoint
	if checkpoint == nil {
		checkpoint = make(map[string]interface{})
	}

	// Determine since timestamp for incremental sync
	sinceTime := io.getSinceTime(checkpoint)

	// Fetch events from connector
	events, err := connector.FetchEvents(ctx, sinceTime, 1000)
	if err != nil {
		io.handleIngestionError(ctx, task, integration, fmt.Errorf("failed to fetch events: %w", err))
		return
	}

	io.logger.Info("Fetched events from platform", map[string]interface{}{
		"integration_id": integration.ID,
		"platform":       integration.Platform,
		"event_count":    len(events),
		"since":          sinceTime,
	})

	// Deduplicate events
	events = io.deduplicateEvents(ctx, integration.ID, events, checkpoint)

	if len(events) == 0 {
		io.logger.Info("No new events to process", map[string]interface{}{
			"integration_id": integration.ID,
		})
		task.Status = TaskStatusCompleted
		io.updateSyncStatus(ctx, integration.ID, "success", nil)
		return
	}

	// Normalize events
	connectorNormalizedEvents, err := connector.NormalizeData(ctx, events)
	if err != nil {
		io.handleIngestionError(ctx, task, integration, fmt.Errorf("failed to normalize events: %w", err))
		return
	}

	// Convert connector normalized events to service normalized events
	normalizedEvents := make([]NormalizedEvent, len(connectorNormalizedEvents))
	for i, ce := range connectorNormalizedEvents {
		normalizedEvents[i] = NormalizedEvent{
			PlatformID:  ce.PlatformID,
			EventType:   EventType(ce.EventType),
			Timestamp:   ce.Timestamp,
			Author:      ce.Author,
			Content:     ce.Content,
			Title:       ce.Title,
			ThreadID:    ce.ThreadID,
			ParentID:    ce.ParentID,
			FileRefs:    ce.FileRefs,
			FeatureRefs: ce.FeatureRefs,
			Labels:      ce.Labels,
			State:       ce.State,
			Metadata:    ce.Metadata,
			Platform:    ce.Platform,
		}
	}

	// Process events through context processor
	if io.contextProcessor != nil {
		processingResult, err := io.contextProcessor.ProcessEvents(ctx, normalizedEvents)
		if err != nil {
			io.logger.Error("Failed to process events through context processor", err, map[string]interface{}{
				"integration_id": integration.ID,
				"event_count":    len(normalizedEvents),
			})
			// Continue even if processing fails - we still want to update checkpoint
		} else {
			io.logger.Info("Processed events through context processor", map[string]interface{}{
				"integration_id":      integration.ID,
				"decision_count":      len(processingResult.DecisionRecords),
				"discussion_count":    len(processingResult.DiscussionSummaries),
				"feature_count":       len(processingResult.FeatureContexts),
				"file_context_count":  len(processingResult.FileContexts),
				"relationship_count":  len(processingResult.Relationships),
			})
		}
	}

	// Update sync checkpoint
	newCheckpoint := io.updateCheckpoint(checkpoint, events, normalizedEvents)
	if err := io.UpdateSyncCheckpoint(ctx, integration.ID, newCheckpoint); err != nil {
		io.logger.Error("Failed to update sync checkpoint", err, map[string]interface{}{
			"integration_id": integration.ID,
		})
	}

	// Update sync status
	now := time.Now()
	task.LastSyncAt = &now
	task.Status = TaskStatusCompleted
	io.updateSyncStatus(ctx, integration.ID, "success", nil)

	io.logger.Info("Completed integration ingestion", map[string]interface{}{
		"integration_id": integration.ID,
		"event_count":    len(events),
		"duration":       time.Since(task.StartedAt).String(),
	})
}

// StartDataSourceIngestion starts ingestion for a specific data source
func (io *IngestionOrchestratorImpl) StartDataSourceIngestion(ctx context.Context, dataSourceID string) error {
	// Get data source
	dataSource, err := io.store.GetProjectDataSource(ctx, dataSourceID)
	if err != nil {
		return fmt.Errorf("failed to get data source: %w", err)
	}

	if !dataSource.IsActive {
		return fmt.Errorf("data source is not active")
	}

	// Start ingestion for the parent integration
	return io.StartIntegrationIngestion(ctx, dataSource.IntegrationID)
}

// StopProjectIngestion stops ingestion for all integrations in a project
func (io *IngestionOrchestratorImpl) StopProjectIngestion(ctx context.Context, projectID string) error {
	io.mu.Lock()
	defer io.mu.Unlock()

	var stoppedCount int
	for _, task := range io.activeIngestions {
		if task.ProjectID == projectID {
			task.Cancel()
			task.Status = TaskStatusPaused
			stoppedCount++
		}
	}

	io.logger.Info("Stopped project ingestion", map[string]interface{}{
		"project_id":    projectID,
		"stopped_count": stoppedCount,
	})

	return nil
}

// GetIngestionHealth gets the health status of project ingestion
func (io *IngestionOrchestratorImpl) GetIngestionHealth(ctx context.Context, projectID string) (*IngestionHealthStatus, error) {
	// Get all integrations for the project
	integrations, err := io.store.GetProjectIntegrations(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project integrations: %w", err)
	}

	health := &IngestionHealthStatus{
		ProjectID:          projectID,
		ActiveIntegrations: 0,
		HealthyIntegrations: 0,
		FailedIntegrations: 0,
		Integrations:       make([]IntegrationHealthStatus, 0),
	}

	var latestSync *time.Time

	for _, integration := range integrations {
		integrationHealth, err := io.GetIntegrationHealth(ctx, integration.ID)
		if err != nil {
			io.logger.Error("Failed to get integration health", err, map[string]interface{}{
				"integration_id": integration.ID,
			})
			continue
		}

		health.Integrations = append(health.Integrations, *integrationHealth)

		if integration.Status == string(models.IntegrationStatusActive) {
			health.ActiveIntegrations++
		}

		if integrationHealth.Status == "healthy" {
			health.HealthyIntegrations++
		} else if integrationHealth.Status == "failed" {
			health.FailedIntegrations++
		}

		if integration.LastSyncAt != nil {
			if latestSync == nil || integration.LastSyncAt.After(*latestSync) {
				latestSync = integration.LastSyncAt
			}
		}
	}

	health.LastSyncAt = latestSync

	// Determine overall status
	if health.FailedIntegrations > 0 {
		health.OverallStatus = "degraded"
	} else if health.HealthyIntegrations == health.ActiveIntegrations {
		health.OverallStatus = "healthy"
	} else {
		health.OverallStatus = "partial"
	}

	return health, nil
}

// GetIntegrationHealth gets the health status of an integration
func (io *IngestionOrchestratorImpl) GetIntegrationHealth(ctx context.Context, integrationID string) (*IntegrationHealthStatus, error) {
	// Get integration
	integration, err := io.store.GetProjectIntegration(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	// Get data sources
	dataSources, err := io.store.GetProjectDataSourcesByIntegration(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get data sources: %w", err)
	}

	// Count active data sources
	activeCount := 0
	for _, ds := range dataSources {
		if ds.IsActive {
			activeCount++
		}
	}

	// Check if task is running
	io.mu.RLock()
	task, isRunning := io.activeIngestions[integrationID]
	io.mu.RUnlock()

	errorCount := 0
	if isRunning {
		errorCount = task.ErrorCount
	}

	// Determine status
	status := "healthy"
	if integration.Status == string(models.IntegrationStatusError) {
		status = "failed"
	} else if integration.Status == string(models.IntegrationStatusInactive) {
		status = "inactive"
	} else if errorCount > 0 {
		status = "degraded"
	}

	health := &IntegrationHealthStatus{
		IntegrationID:     integrationID,
		Platform:          integration.Platform,
		Status:            status,
		LastSyncAt:        integration.LastSyncAt,
		LastSyncStatus:    integration.LastSyncStatus,
		ErrorMessage:      integration.ErrorMessage,
		ErrorCount:        errorCount,
		DataSourceCount:   len(dataSources),
		ActiveDataSources: activeCount,
		SyncCheckpoint:    integration.SyncCheckpoint,
	}

	// Calculate next sync time if applicable
	if isRunning && task.LastSyncAt != nil {
		nextSync := task.LastSyncAt.Add(io.healthCheckInterval)
		health.NextSyncScheduled = &nextSync
	}

	return health, nil
}

// GetSyncCheckpoint gets the sync checkpoint for an integration
func (io *IngestionOrchestratorImpl) GetSyncCheckpoint(ctx context.Context, integrationID string) (map[string]interface{}, error) {
	integration, err := io.store.GetProjectIntegration(ctx, integrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	if integration.SyncCheckpoint == nil {
		return make(map[string]interface{}), nil
	}

	return integration.SyncCheckpoint, nil
}

// UpdateSyncCheckpoint updates the sync checkpoint for an integration
func (io *IngestionOrchestratorImpl) UpdateSyncCheckpoint(ctx context.Context, integrationID string, checkpoint map[string]interface{}) error {
	updates := map[string]interface{}{
		"sync_checkpoint": checkpoint,
		"updated_at":      time.Now(),
	}

	if err := io.store.UpdateProjectIntegration(ctx, integrationID, updates); err != nil {
		return fmt.Errorf("failed to update sync checkpoint: %w", err)
	}

	io.logger.Info("Updated sync checkpoint", map[string]interface{}{
		"integration_id": integrationID,
		"checkpoint":     checkpoint,
	})

	return nil
}

// RetryFailedIngestion retries a failed ingestion
func (io *IngestionOrchestratorImpl) RetryFailedIngestion(ctx context.Context, integrationID string) error {
	// Get integration
	_, err := io.store.GetProjectIntegration(ctx, integrationID)
	if err != nil {
		return fmt.Errorf("failed to get integration: %w", err)
	}

	// Reset error state
	updates := map[string]interface{}{
		"status":           string(models.IntegrationStatusActive),
		"error_message":    nil,
		"last_sync_status": nil,
		"updated_at":       time.Now(),
	}

	if err := io.store.UpdateProjectIntegration(ctx, integrationID, updates); err != nil {
		return fmt.Errorf("failed to reset integration status: %w", err)
	}

	// Start ingestion
	return io.StartIntegrationIngestion(ctx, integrationID)
}

// StartOrchestrator starts the background orchestrator
func (io *IngestionOrchestratorImpl) StartOrchestrator(ctx context.Context) error {
	io.mu.Lock()
	defer io.mu.Unlock()

	if io.orchestratorCtx != nil {
		return fmt.Errorf("orchestrator already running")
	}

	io.orchestratorCtx, io.orchestratorCancel = context.WithCancel(ctx)

	// Start health check loop
	io.orchestratorWg.Add(1)
	go io.healthCheckLoop()

	io.logger.Info("Started ingestion orchestrator", nil)

	return nil
}

// StopOrchestrator stops the background orchestrator
func (io *IngestionOrchestratorImpl) StopOrchestrator(ctx context.Context) error {
	io.mu.Lock()
	if io.orchestratorCancel != nil {
		io.orchestratorCancel()
	}
	io.mu.Unlock()

	// Wait for background tasks to complete
	io.orchestratorWg.Wait()

	io.logger.Info("Stopped ingestion orchestrator", nil)

	return nil
}

// healthCheckLoop performs periodic health checks and retries
func (io *IngestionOrchestratorImpl) healthCheckLoop() {
	defer io.orchestratorWg.Done()

	ticker := time.NewTicker(io.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-io.orchestratorCtx.Done():
			return
		case <-ticker.C:
			io.performHealthCheck()
		}
	}
}

// performHealthCheck performs health checks on all active integrations
func (io *IngestionOrchestratorImpl) performHealthCheck() {
	io.mu.RLock()
	tasks := make([]*IngestionTask, 0, len(io.activeIngestions))
	for _, task := range io.activeIngestions {
		tasks = append(tasks, task)
	}
	io.mu.RUnlock()

	for _, task := range tasks {
		// Check if task has been running too long without sync
		if task.LastSyncAt == nil && time.Since(task.StartedAt) > time.Hour {
			io.logger.Error("Ingestion task running too long without sync", nil, map[string]interface{}{
				"integration_id": task.IntegrationID,
				"duration":       time.Since(task.StartedAt).String(),
			})
		}

		// Check if task has too many errors
		if task.ErrorCount >= io.maxRetries {
			io.logger.Error("Ingestion task exceeded max retries", nil, map[string]interface{}{
				"integration_id": task.IntegrationID,
				"error_count":    task.ErrorCount,
			})
			task.Cancel()
			task.Status = TaskStatusFailed
		}
	}
}

// Helper methods

// handleIngestionError handles errors during ingestion
func (io *IngestionOrchestratorImpl) handleIngestionError(ctx context.Context, task *IngestionTask, integration *models.ProjectIntegration, err error) {
	task.ErrorCount++
	errorMsg := err.Error()
	task.LastError = &errorMsg

	io.logger.Error("Ingestion error", err, map[string]interface{}{
		"integration_id": task.IntegrationID,
		"error_count":    task.ErrorCount,
	})

	// Update integration status
	status := string(models.IntegrationStatusActive)
	if task.ErrorCount >= io.maxRetries {
		status = string(models.IntegrationStatusError)
		task.Status = TaskStatusFailed
	}

	io.updateSyncStatus(ctx, integration.ID, "failed", &errorMsg)

	updates := map[string]interface{}{
		"status":        status,
		"error_message": errorMsg,
		"updated_at":    time.Now(),
	}

	if err := io.store.UpdateProjectIntegration(ctx, integration.ID, updates); err != nil {
		io.logger.Error("Failed to update integration error status", err, map[string]interface{}{
			"integration_id": integration.ID,
		})
	}

	// Schedule retry if not exceeded max retries
	if task.ErrorCount < io.maxRetries {
		backoff := io.retryBackoff * time.Duration(task.ErrorCount)
		io.logger.Info("Scheduling retry", map[string]interface{}{
			"integration_id": task.IntegrationID,
			"backoff":        backoff.String(),
			"attempt":        task.ErrorCount + 1,
		})

		time.AfterFunc(backoff, func() {
			if err := io.StartIntegrationIngestion(context.Background(), task.IntegrationID); err != nil {
				io.logger.Error("Failed to retry ingestion", err, map[string]interface{}{
					"integration_id": task.IntegrationID,
				})
			}
		})
	}
}

// updateSyncStatus updates the sync status of an integration
func (io *IngestionOrchestratorImpl) updateSyncStatus(ctx context.Context, integrationID, status string, errorMsg *string) {
	now := time.Now()
	updates := map[string]interface{}{
		"last_sync_at":     now,
		"last_sync_status": status,
		"updated_at":       now,
	}

	if errorMsg != nil {
		updates["error_message"] = *errorMsg
	} else {
		updates["error_message"] = nil
	}

	if err := io.store.UpdateProjectIntegration(ctx, integrationID, updates); err != nil {
		io.logger.Error("Failed to update sync status", err, map[string]interface{}{
			"integration_id": integrationID,
		})
	}
}

// getSinceTime determines the since timestamp for incremental sync
func (io *IngestionOrchestratorImpl) getSinceTime(checkpoint map[string]interface{}) time.Time {
	if lastSync, ok := checkpoint["last_sync_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastSync); err == nil {
			return t
		}
	}

	// Default to 24 hours ago
	return time.Now().Add(-24 * time.Hour)
}

// deduplicateEvents removes duplicate events based on checkpoint
func (io *IngestionOrchestratorImpl) deduplicateEvents(ctx context.Context, integrationID string, events []PlatformEvent, checkpoint map[string]interface{}) []PlatformEvent {
	// Get processed event IDs from checkpoint
	processedIDs := make(map[string]bool)
	if ids, ok := checkpoint["processed_event_ids"].([]interface{}); ok {
		for _, id := range ids {
			if idStr, ok := id.(string); ok {
				processedIDs[idStr] = true
			}
		}
	}

	// Filter out already processed events
	var deduplicated []PlatformEvent
	for _, event := range events {
		if !processedIDs[event.ID] {
			deduplicated = append(deduplicated, event)
		}
	}

	io.logger.Info("Deduplicated events", map[string]interface{}{
		"integration_id":   integrationID,
		"original_count":   len(events),
		"deduplicated_count": len(deduplicated),
		"filtered_count":   len(events) - len(deduplicated),
	})

	return deduplicated
}

// updateCheckpoint updates the checkpoint with new event information
func (io *IngestionOrchestratorImpl) updateCheckpoint(checkpoint map[string]interface{}, events []PlatformEvent, normalizedEvents []NormalizedEvent) map[string]interface{} {
	if checkpoint == nil {
		checkpoint = make(map[string]interface{})
	}

	// Update last sync time
	checkpoint["last_sync_time"] = time.Now().Format(time.RFC3339)

	// Track processed event IDs (keep only recent ones within deduplication window)
	processedIDs := make([]string, 0)
	cutoffTime := time.Now().Add(-io.deduplicationWindow)

	// Add new event IDs
	for _, event := range events {
		if event.Timestamp.After(cutoffTime) {
			processedIDs = append(processedIDs, event.ID)
		}
	}

	// Keep existing IDs that are still within window
	if existingIDs, ok := checkpoint["processed_event_ids"].([]interface{}); ok {
		for _, id := range existingIDs {
			if idStr, ok := id.(string); ok {
				// Only keep if not already in new list
				found := false
				for _, newID := range processedIDs {
					if newID == idStr {
						found = true
						break
					}
				}
				if !found && len(processedIDs) < 10000 { // Limit to 10k IDs
					processedIDs = append(processedIDs, idStr)
				}
			}
		}
	}

	checkpoint["processed_event_ids"] = processedIDs

	// Update event counts
	checkpoint["total_events_processed"] = getIntValue(checkpoint, "total_events_processed") + len(events)
	checkpoint["last_batch_size"] = len(events)

	// Track latest event timestamp
	if len(normalizedEvents) > 0 {
		latestTimestamp := normalizedEvents[0].Timestamp
		for _, event := range normalizedEvents {
			if event.Timestamp.After(latestTimestamp) {
				latestTimestamp = event.Timestamp
			}
		}
		checkpoint["latest_event_timestamp"] = latestTimestamp.Format(time.RFC3339)
	}

	return checkpoint
}

// getIntValue safely gets an int value from checkpoint
func getIntValue(checkpoint map[string]interface{}, key string) int {
	if val, ok := checkpoint[key].(float64); ok {
		return int(val)
	}
	if val, ok := checkpoint[key].(int); ok {
		return val
	}
	return 0
}
