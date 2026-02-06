package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/contextkeeper/internal/config"
	"github.com/yourusername/contextkeeper/internal/repository"
	"github.com/yourusername/contextkeeper/internal/server"
)

// TestResponseTimeRequirements validates that response times meet requirements
func TestResponseTimeRequirements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cfg := getTestConfig()
	repo := repository.NewRepository(cfg)
	defer repo.Close()

	srv := server.NewServer(cfg, repo)
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	token, projectID := setupUserAndProject(t, testServer.URL)

	// Populate with test data
	populateLargeDataset(t, testServer.URL, projectID, token)

	t.Run("MCP Tool Response Times", func(t *testing.T) {
		tests := []struct {
			name          string
			endpoint      string
			maxDuration   time.Duration
			description   string
		}{
			{
				name:        "Search Project Knowledge",
				endpoint:    "/api/mcp/search_project_knowledge",
				maxDuration: 1 * time.Second,
				description: "Search queries should complete within 1 second",
			},
			{
				name:        "Get File Context",
				endpoint:    "/api/mcp/get_context_for_file",
				maxDuration: 500 * time.Millisecond,
				description: "File context retrieval should complete within 500ms",
			},
			{
				name:        "Get Decision History",
				endpoint:    "/api/mcp/get_decision_history",
				maxDuration: 1 * time.Second,
				description: "Decision history should complete within 1 second",
			},
			{
				name:        "List Recent Discussions",
				endpoint:    "/api/mcp/list_recent_architecture_discussions",
				maxDuration: 800 * time.Millisecond,
				description: "Recent discussions should complete within 800ms",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Warm up cache
				makeRequest(t, testServer.URL+tt.endpoint, "POST", map[string]string{"query": "test"}, token)

				// Measure response time
				start := time.Now()
				resp := makeRequest(t, testServer.URL+tt.endpoint, "POST", map[string]string{"query": "authentication"}, token)
				duration := time.Since(start)

				if resp.StatusCode != http.StatusOK {
					t.Errorf("Request failed with status %d", resp.StatusCode)
				}

				if duration > tt.maxDuration {
					t.Errorf("%s: Response time %v exceeds maximum %v", tt.description, duration, tt.maxDuration)
				} else {
					t.Logf("✓ %s: %v (max: %v)", tt.name, duration, tt.maxDuration)
				}
			})
		}
	})

	t.Run("Caching Performance Improvement", func(t *testing.T) {
		endpoint := "/api/mcp/search_project_knowledge"
		query := map[string]string{"query": "authentication implementation"}

		// First request (cold cache)
		start := time.Now()
		makeRequest(t, testServer.URL+endpoint, "POST", query, token)
		coldDuration := time.Since(start)

		// Second request (warm cache)
		start = time.Now()
		makeRequest(t, testServer.URL+endpoint, "POST", query, token)
		warmDuration := time.Since(start)

		// Cached request should be at least 30% faster
		improvement := float64(coldDuration-warmDuration) / float64(coldDuration) * 100

		if improvement < 30 {
			t.Errorf("Cache improvement %.1f%% is less than expected 30%%", improvement)
		} else {
			t.Logf("✓ Cache improved performance by %.1f%% (cold: %v, warm: %v)", improvement, coldDuration, warmDuration)
		}
	})
}

// TestLoadHandling tests system behavior under realistic load
func TestLoadHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	cfg := getTestConfig()
	repo := repository.NewRepository(cfg)
	defer repo.Close()

	srv := server.NewServer(cfg, repo)
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	token, projectID := setupUserAndProject(t, testServer.URL)
	populateLargeDataset(t, testServer.URL, projectID, token)

	t.Run("Concurrent User Load", func(t *testing.T) {
		concurrentUsers := 50
		requestsPerUser := 10
		totalRequests := concurrentUsers * requestsPerUser

		var successCount int64
		var errorCount int64
		var totalDuration int64

		var wg sync.WaitGroup
		startTime := time.Now()

		for i := 0; i < concurrentUsers; i++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				for j := 0; j < requestsPerUser; j++ {
					reqStart := time.Now()
					resp := makeRequest(t, testServer.URL+"/api/mcp/search_project_knowledge", "POST", 
						map[string]string{"query": fmt.Sprintf("query-%d-%d", userID, j)}, token)
					reqDuration := time.Since(reqStart)

					atomic.AddInt64(&totalDuration, int64(reqDuration))

					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&successCount, 1)
					} else {
						atomic.AddInt64(&errorCount, 1)
					}
				}
			}(i)
		}

		wg.Wait()
		totalTime := time.Since(startTime)

		successRate := float64(successCount) / float64(totalRequests) * 100
		avgDuration := time.Duration(totalDuration / int64(totalRequests))
		throughput := float64(totalRequests) / totalTime.Seconds()

		t.Logf("Load Test Results:")
		t.Logf("  Total Requests: %d", totalRequests)
		t.Logf("  Successful: %d (%.1f%%)", successCount, successRate)
		t.Logf("  Failed: %d", errorCount)
		t.Logf("  Total Time: %v", totalTime)
		t.Logf("  Avg Response Time: %v", avgDuration)
		t.Logf("  Throughput: %.1f req/s", throughput)

		// Validate requirements
		if successRate < 99.0 {
			t.Errorf("Success rate %.1f%% is below required 99%%", successRate)
		}

		if avgDuration > 2*time.Second {
			t.Errorf("Average response time %v exceeds acceptable limit", avgDuration)
		}

		if throughput < 10 {
			t.Errorf("Throughput %.1f req/s is below minimum 10 req/s", throughput)
		}

		t.Log("✅ System handled concurrent load successfully")
	})

	t.Run("Sustained Load Test", func(t *testing.T) {
		duration := 30 * time.Second
		requestRate := 20 // requests per second
		interval := time.Second / time.Duration(requestRate)

		var successCount int64
		var errorCount int64
		var responseTimes []time.Duration
		var mu sync.Mutex

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		done := time.After(duration)
		startTime := time.Now()

		t.Logf("Starting sustained load test: %d req/s for %v", requestRate, duration)

	loop:
		for {
			select {
			case <-ticker.C:
				go func() {
					reqStart := time.Now()
					resp := makeRequest(t, testServer.URL+"/api/mcp/search_project_knowledge", "POST",
						map[string]string{"query": "sustained load test"}, token)
					reqDuration := time.Since(reqStart)

					mu.Lock()
					responseTimes = append(responseTimes, reqDuration)
					mu.Unlock()

					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&successCount, 1)
					} else {
						atomic.AddInt64(&errorCount, 1)
					}
				}()
			case <-done:
				break loop
			}
		}

		// Wait for in-flight requests
		time.Sleep(2 * time.Second)

		totalRequests := successCount + errorCount
		actualDuration := time.Since(startTime)
		successRate := float64(successCount) / float64(totalRequests) * 100
		actualThroughput := float64(totalRequests) / actualDuration.Seconds()

		// Calculate percentiles
		mu.Lock()
		p50, p95, p99 := calculatePercentiles(responseTimes)
		mu.Unlock()

		t.Logf("Sustained Load Results:")
		t.Logf("  Duration: %v", actualDuration)
		t.Logf("  Total Requests: %d", totalRequests)
		t.Logf("  Success Rate: %.1f%%", successRate)
		t.Logf("  Throughput: %.1f req/s", actualThroughput)
		t.Logf("  Response Times:")
		t.Logf("    P50: %v", p50)
		t.Logf("    P95: %v", p95)
		t.Logf("    P99: %v", p99)

		// Validate sustained performance
		if successRate < 99.0 {
			t.Errorf("Success rate %.1f%% dropped below 99%% under sustained load", successRate)
		}

		if p95 > 2*time.Second {
			t.Errorf("P95 response time %v exceeds 2 seconds", p95)
		}

		t.Log("✅ System maintained performance under sustained load")
	})
}

// TestGracefulDegradation tests system behavior under stress
func TestGracefulDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	cfg := getTestConfig()
	repo := repository.NewRepository(cfg)
	defer repo.Close()

	srv := server.NewServer(cfg, repo)
	testServer := httptest.NewServer(srv.Router)
	defer testServer.Close()

	token, projectID := setupUserAndProject(t, testServer.URL)

	t.Run("High Memory Pressure", func(t *testing.T) {
		// Simulate high memory usage by requesting large datasets
		largeQuery := map[string]interface{}{
			"query":      "large dataset",
			"limit":      10000,
			"include_all": true,
		}

		resp := makeRequest(t, testServer.URL+"/api/mcp/search_project_knowledge", "POST", largeQuery, token)

		// System should either:
		// 1. Return partial results with a warning
		// 2. Return 503 Service Unavailable
		// 3. Return results within reasonable time

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

			// Check if system limited results
			if results, ok := result["results"].([]interface{}); ok {
				if len(results) < 10000 {
					t.Logf("✓ System gracefully limited results to %d items", len(results))
				}
			}

			// Check for warning message
			if warning, ok := result["warning"].(string); ok {
				t.Logf("✓ System provided warning: %s", warning)
			}
		} else if resp.StatusCode == http.StatusServiceUnavailable {
			t.Log("✓ System returned 503 under high load (acceptable)")
		} else {
			t.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
	})

	t.Run("Database Connection Pool Exhaustion", func(t *testing.T) {
		// Simulate many concurrent database-heavy operations
		concurrentRequests := 100
		var wg sync.WaitGroup
		var successCount int64
		var serviceUnavailableCount int64

		for i := 0; i < concurrentRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				resp := makeRequest(t, testServer.URL+"/api/projects/"+projectID+"/data", "GET", nil, token)
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else if resp.StatusCode == http.StatusServiceUnavailable {
					atomic.AddInt64(&serviceUnavailableCount, 1)
				}
			}()
		}

		wg.Wait()

		totalHandled := successCount + serviceUnavailableCount
		if totalHandled < int64(concurrentRequests) {
			t.Errorf("System failed to handle %d requests", concurrentRequests-int(totalHandled))
		}

		if serviceUnavailableCount > 0 {
			t.Logf("✓ System gracefully returned 503 for %d requests under connection pressure", serviceUnavailableCount)
		}

		if successCount > 0 {
			t.Logf("✓ System successfully handled %d requests", successCount)
		}
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		// Send requests rapidly to trigger rate limiting
		rapidRequests := 200
		var rateLimitedCount int64

		for i := 0; i < rapidRequests; i++ {
			resp := makeRequest(t, testServer.URL+"/api/mcp/search_project_knowledge", "POST",
				map[string]string{"query": "rate limit test"}, token)

			if resp.StatusCode == http.StatusTooManyRequests {
				atomic.AddInt64(&rateLimitedCount, 1)
			}
		}

		if rateLimitedCount > 0 {
			t.Logf("✓ Rate limiting activated after %d requests", rapidRequests-int(rateLimitedCount))
		} else {
			t.Log("⚠ Rate limiting not triggered (may need adjustment)")
		}
	})
}

// TestHorizontalScaling tests multi-instance deployment
func TestHorizontalScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scaling test in short mode")
	}

	t.Run("Multiple Server Instances", func(t *testing.T) {
		cfg := getTestConfig()

		// Start multiple server instances
		numInstances := 3
		servers := make([]*httptest.Server, numInstances)

		for i := 0; i < numInstances; i++ {
			repo := repository.NewRepository(cfg)
			srv := server.NewServer(cfg, repo)
			servers[i] = httptest.NewServer(srv.Router)
			defer servers[i].Close()
			defer repo.Close()
		}

		// Create user and project on first instance
		token, projectID := setupUserAndProject(t, servers[0].URL)

		// Distribute requests across all instances
		var wg sync.WaitGroup
		requestsPerInstance := 20
		var successCount int64

		for i := 0; i < numInstances; i++ {
			wg.Add(1)
			go func(instanceURL string) {
				defer wg.Done()

				for j := 0; j < requestsPerInstance; j++ {
					resp := makeRequest(t, instanceURL+"/api/mcp/search_project_knowledge", "POST",
						map[string]string{"query": "scaling test"}, token)

					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&successCount, 1)
					}
				}
			}(servers[i].URL)
		}

		wg.Wait()

		totalRequests := int64(numInstances * requestsPerInstance)
		successRate := float64(successCount) / float64(totalRequests) * 100

		if successRate < 95 {
			t.Errorf("Success rate %.1f%% is too low for horizontal scaling", successRate)
		}

		t.Logf("✓ Horizontal scaling: %d instances handled %d requests with %.1f%% success rate",
			numInstances, totalRequests, successRate)
	})
}

// Helper functions

func populateLargeDataset(t *testing.T, baseURL, projectID, token string) {
	t.Log("Populating large dataset for performance testing...")

	// Add multiple integrations
	platforms := []string{"github", "slack", "discord"}
	for _, platform := range platforms {
		integrationData := map[string]interface{}{
			"platform":     platform,
			"access_token": "test-token",
			"config":       map[string]interface{}{},
		}
		makeRequest(t, baseURL+"/api/projects/"+projectID+"/integrations", "POST", integrationData, token)
	}

	// Simulate data ingestion (in real scenario, this would be done by background workers)
	// For testing, we'll add mock data directly
	for i := 0; i < 100; i++ {
		data := map[string]interface{}{
			"type":    "test_event",
			"content": fmt.Sprintf("Test event %d with searchable content", i),
		}
		makeRequest(t, baseURL+"/api/projects/"+projectID+"/data", "POST", data, token)
	}

	t.Log("✓ Dataset populated")
}

func calculatePercentiles(durations []time.Duration) (p50, p95, p99 time.Duration) {
	if len(durations) == 0 {
		return 0, 0, 0
	}

	// Sort durations
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	// Simple bubble sort (good enough for test data)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	p50 = sorted[len(sorted)*50/100]
	p95 = sorted[len(sorted)*95/100]
	p99 = sorted[len(sorted)*99/100]

	return p50, p95, p99
}
