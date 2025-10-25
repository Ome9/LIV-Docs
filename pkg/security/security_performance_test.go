// Performance and stress tests for security and administration systems
// Tests system performance under load and stress conditions

package security

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liv-format/liv/pkg/core"
)

// BenchmarkPermissionEvaluation benchmarks permission evaluation performance
func BenchmarkPermissionEvaluation(b *testing.B) {
	tempDir, err := ioutil.TempDir("", "perf-test-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	// Setup
	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{DefaultPolicyID: "default"}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	mockSM := &MockSecurityManager{}
	mockCP := &MockCryptoProvider{}
	mockLogger := &MockLogger{}

	permManager := NewPermissionManager(pm, mockSM, mockCP, mockLogger)

	// Create test policy
	policy := createTestPolicy("bench-policy", "Benchmark Policy")
	err = pm.CreatePolicy(context.Background(), policy, "admin")
	require.NoError(b, err)

	// Setup mock expectations
	mockSM.On("EvaluatePermissions", 
		&core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024,
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		}, 
		policy.SecurityPolicy).Return(true)
	mockLogger.On("Info", "Permission evaluation completed", 
		"document_id", "bench-doc",
		"policy_id", "bench-policy",
		"granted", true,
		"warnings", 0,
	).Return()

	request := &PermissionRequest{
		DocumentID: "bench-doc",
		RequestedPerms: &core.WASMPermissions{
			MemoryLimit:     8 * 1024 * 1024,
			CPUTimeLimit:    3000,
			AllowNetworking: false,
			AllowFileSystem: false,
			AllowedImports:  []string{"console"},
		},
		PolicyID:      "bench-policy",
		UserContext:   &UserContext{UserID: "bench-user"},
		Justification: "Benchmark test",
		RequestedAt:   time.Now(),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := permManager.EvaluatePermissionRequest(context.Background(), request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkPolicyCreation benchmarks policy creation performance
func BenchmarkPolicyCreation(b *testing.B) {
	tempDir, err := ioutil.TempDir("", "policy-perf-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{DefaultPolicyID: "default"}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy := createTestPolicy(fmt.Sprintf("bench-policy-%d", i), fmt.Sprintf("Benchmark Policy %d", i))
		err := pm.CreatePolicy(context.Background(), policy, "admin")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkResourceMonitoring benchmarks resource monitoring performance
func BenchmarkResourceMonitoring(b *testing.B) {
	tempDir, err := ioutil.TempDir("", "resource-perf-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{DefaultPolicyID: "default"}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	// Create test policy
	policy := createTestPolicy("resource-policy", "Resource Policy")
	err = pm.CreatePolicy(context.Background(), policy, "admin")
	require.NoError(b, err)

	metrics := &ResourceMetrics{
		MemoryUsage:         16 * 1024 * 1024,
		CPUTime:             2000,
		ConcurrentDocuments: 3,
		NetworkBandwidth:    1024 * 1024,
		StorageUsage:        50 * 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.MonitorResourceUsage(context.Background(), metrics)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEventLogging benchmarks security event logging performance
func BenchmarkEventLogging(b *testing.B) {
	tempDir, err := ioutil.TempDir("", "event-perf-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))

	event := &SecurityEvent{
		ID:          "bench-event",
		Timestamp:   time.Now(),
		EventType:   EventPolicyViolation,
		Severity:    SeverityMedium,
		Source:      "benchmark",
		Description: "Benchmark security event",
		Details:     map[string]interface{}{"test": "value"},
		UserID:      "bench-user",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := eventLogger.LogSecurityEvent(event)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestConcurrentPolicyOperations tests concurrent policy operations
func TestConcurrentPolicyOperations(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "concurrent-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{
		DefaultPolicyID:         "default",
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
	}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	ctx := context.Background()
	numGoroutines := 10
	numOperations := 100

	// Test concurrent policy creation
	t.Run("ConcurrentPolicyCreation", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numOperations)

		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for i := 0; i < numOperations; i++ {
					policyID := fmt.Sprintf("concurrent-policy-%d-%d", goroutineID, i)
					policy := createTestPolicy(policyID, fmt.Sprintf("Concurrent Policy %d-%d", goroutineID, i))
					
					err := pm.CreatePolicy(ctx, policy, fmt.Sprintf("user-%d", goroutineID))
					if err != nil {
						errors <- err
						return
					}
				}
			}(g)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent policy creation error: %v", err)
		}

		// Verify all policies were created
		policies, err := pm.ListPolicies(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(policies), numGoroutines*numOperations, "Should have created all policies")
	})

	// Test concurrent permission evaluations
	t.Run("ConcurrentPermissionEvaluations", func(t *testing.T) {
		mockSM := &MockSecurityManager{}
		mockCP := &MockCryptoProvider{}
		mockLogger := &MockLogger{}

		permManager := NewPermissionManager(pm, mockSM, mockCP, mockLogger)

		// Create test policy
		testPolicy := createTestPolicy("eval-policy", "Evaluation Policy")
		err := pm.CreatePolicy(ctx, testPolicy, "admin")
		require.NoError(t, err)

		// Setup mock expectations for concurrent calls
		mockSM.On("EvaluatePermissions", 
			&core.WASMPermissions{
				MemoryLimit:     8 * 1024 * 1024,
				CPUTimeLimit:    3000,
				AllowNetworking: false,
				AllowFileSystem: false,
				AllowedImports:  []string{"console"},
			}, 
			testPolicy.SecurityPolicy).Return(true)
		mockLogger.On("Info", "Permission evaluation completed", 
			"document_id", "eval-doc",
			"policy_id", "eval-policy",
			"granted", true,
			"warnings", 0,
		).Return()

		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*numOperations)

		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for i := 0; i < numOperations; i++ {
					request := &PermissionRequest{
						DocumentID: "eval-doc",
						RequestedPerms: &core.WASMPermissions{
							MemoryLimit:     8 * 1024 * 1024,
							CPUTimeLimit:    3000,
							AllowNetworking: false,
							AllowFileSystem: false,
							AllowedImports:  []string{"console"},
						},
						PolicyID:      "eval-policy",
						UserContext:   &UserContext{UserID: fmt.Sprintf("user-%d", goroutineID)},
						Justification: fmt.Sprintf("Concurrent test %d-%d", goroutineID, i),
						RequestedAt:   time.Now(),
					}

					_, err := permManager.EvaluatePermissionRequest(ctx, request)
					if err != nil {
						errors <- err
						return
					}
				}
			}(g)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent permission evaluation error: %v", err)
		}
	})
}

// TestMemoryUsageUnderLoad tests memory usage under load
func TestMemoryUsageUnderLoad(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "memory-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Force garbage collection before test
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{DefaultPolicyID: "default"}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	ctx := context.Background()

	// Create many policies to test memory usage
	numPolicies := 1000
	for i := 0; i < numPolicies; i++ {
		policy := createTestPolicy(fmt.Sprintf("memory-policy-%d", i), fmt.Sprintf("Memory Policy %d", i))
		err := pm.CreatePolicy(ctx, policy, "admin")
		require.NoError(t, err)
	}

	// Force garbage collection after test
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Calculate memory usage
	memoryUsed := m2.Alloc - m1.Alloc
	memoryPerPolicy := memoryUsed / uint64(numPolicies)

	t.Logf("Memory usage: %d bytes total, %d bytes per policy", memoryUsed, memoryPerPolicy)

	// Memory usage should be reasonable (less than 10KB per policy)
	assert.Less(t, memoryPerPolicy, uint64(10*1024), "Memory usage per policy should be reasonable")
}

// TestEventLogPerformanceUnderLoad tests event logging performance under high load
func TestEventLogPerformanceUnderLoad(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "event-load-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))

	numEvents := 10000
	numGoroutines := 10
	eventsPerGoroutine := numEvents / numGoroutines

	start := time.Now()

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < eventsPerGoroutine; i++ {
				event := &SecurityEvent{
					ID:          fmt.Sprintf("load-event-%d-%d", goroutineID, i),
					Timestamp:   time.Now(),
					EventType:   EventPolicyViolation,
					Severity:    SeverityMedium,
					Source:      "load-test",
					Description: fmt.Sprintf("Load test event %d-%d", goroutineID, i),
					Details:     map[string]interface{}{"goroutine": goroutineID, "event": i},
					UserID:      fmt.Sprintf("user-%d", goroutineID),
				}

				err := eventLogger.LogSecurityEvent(event)
				if err != nil {
					errors <- err
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)

	// Check for errors
	for err := range errors {
		t.Errorf("Event logging error: %v", err)
	}

	// Calculate performance metrics
	eventsPerSecond := float64(numEvents) / duration.Seconds()
	t.Logf("Event logging performance: %d events in %v (%.2f events/sec)", numEvents, duration, eventsPerSecond)

	// Should be able to log at least 1000 events per second
	assert.Greater(t, eventsPerSecond, 1000.0, "Event logging should handle at least 1000 events/sec")

	// Verify all events were logged
	events, err := eventLogger.GetSecurityEvents(&EventFilter{})
	assert.NoError(t, err)
	assert.Len(t, events, numEvents, "Should have logged all events")
}

// TestSystemValidationPerformance tests system validation performance
func TestSystemValidationPerformance(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "validation-perf-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{
		DefaultPolicyID:         "default",
		EnablePolicyInheritance: true,
		MaxPolicyDepth:          5,
	}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	ctx := context.Background()

	// Create many policies with various configurations
	numPolicies := 500
	for i := 0; i < numPolicies; i++ {
		policy := createTestPolicy(fmt.Sprintf("validation-policy-%d", i), fmt.Sprintf("Validation Policy %d", i))
		
		// Vary policy configurations to test different validation paths
		if i%5 == 0 {
			policy.AdminControls.RequireSignature = true
		}
		if i%7 == 0 {
			policy.EventConfig.EnableAuditLog = false // This should trigger validation issues
		}
		if i%3 == 0 {
			policy.SecurityPolicy.WASMPermissions.MemoryLimit = 128 * 1024 * 1024 // Overly permissive
		}

		err := pm.CreatePolicy(ctx, policy, "admin")
		require.NoError(t, err)
	}

	// Benchmark system validation
	start := time.Now()
	
	numValidations := 100
	for i := 0; i < numValidations; i++ {
		report, err := pm.ValidateSystemConfiguration(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, report)
	}
	
	duration := time.Since(start)
	validationsPerSecond := float64(numValidations) / duration.Seconds()

	t.Logf("System validation performance: %d validations in %v (%.2f validations/sec)", 
		numValidations, duration, validationsPerSecond)

	// Should be able to perform at least 10 validations per second
	assert.Greater(t, validationsPerSecond, 10.0, "System validation should handle at least 10 validations/sec")
}

// TestLargeDocumentProcessing tests processing of large documents
func TestLargeDocumentProcessing(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "large-doc-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{DefaultPolicyID: "default"}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	ctx := context.Background()

	// Create policy for large documents
	largeDocPolicy := createTestPolicy("large-doc-policy", "Large Document Policy")
	largeDocPolicy.AdminControls.MaxDocumentSize = 100 * 1024 * 1024 // 100MB
	largeDocPolicy.ResourceLimits.MaxMemoryPerDocument = 200 * 1024 * 1024 // 200MB
	err = pm.CreatePolicy(ctx, largeDocPolicy, "admin")
	require.NoError(t, err)

	// Create large test document
	largeDoc := createTestDocument()
	
	// Add large content
	largeHTML := make([]byte, 10*1024*1024) // 10MB HTML
	for i := range largeHTML {
		largeHTML[i] = byte('A' + (i % 26))
	}
	largeDoc.Content.HTML = string(largeHTML)

	// Add large WASM modules
	for i := 0; i < 5; i++ {
		moduleData := make([]byte, 5*1024*1024) // 5MB per module
		for j := range moduleData {
			moduleData[j] = byte(i)
		}
		largeDoc.WASMModules[fmt.Sprintf("large-module-%d", i)] = moduleData
	}

	userContext := &UserContext{UserID: "large-doc-user"}

	// Test document evaluation performance
	start := time.Now()
	
	evaluation, err := pm.EvaluateDocumentSecurity(ctx, largeDoc, "large-doc-policy", userContext)
	
	duration := time.Since(start)

	assert.NoError(t, err, "Should evaluate large document successfully")
	assert.NotNil(t, evaluation, "Should return evaluation for large document")

	t.Logf("Large document evaluation time: %v", duration)

	// Should complete evaluation in reasonable time (< 5 seconds)
	assert.Less(t, duration, 5*time.Second, "Large document evaluation should complete quickly")
}

// TestResourceMonitoringAccuracy tests accuracy of resource monitoring
func TestResourceMonitoringAccuracy(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "resource-accuracy-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	eventLogger := NewFileSecurityEventLogger(filepath.Join(tempDir, "security.log"))
	auditLogger := NewFileAuditLogger(filepath.Join(tempDir, "audit.log"))
	config := &PolicyManagerConfig{DefaultPolicyID: "default"}
	pm := NewPolicyManager(config, eventLogger, auditLogger)

	ctx := context.Background()

	// Create precise policy limits
	precisePolicy := createTestPolicy("precise-policy", "Precise Policy")
	precisePolicy.ResourceLimits.MaxMemoryPerDocument = 32 * 1024 * 1024    // Exactly 32MB
	precisePolicy.ResourceLimits.MaxCPUTimePerDocument = 5000               // Exactly 5 seconds
	precisePolicy.ResourceLimits.MaxConcurrentDocuments = 3                 // Exactly 3 documents
	err = pm.CreatePolicy(ctx, precisePolicy, "admin")
	require.NoError(t, err)

	// Test boundary conditions
	testCases := []struct {
		name            string
		metrics         *ResourceMetrics
		expectViolation bool
		description     string
	}{
		{
			name: "ExactlyAtLimit",
			metrics: &ResourceMetrics{
				MemoryUsage:         32 * 1024 * 1024,
				CPUTime:             5000,
				ConcurrentDocuments: 3,
			},
			expectViolation: false,
			description:     "Should not violate when exactly at limit",
		},
		{
			name: "JustOverLimit",
			metrics: &ResourceMetrics{
				MemoryUsage:         32*1024*1024 + 1,
				CPUTime:             5001,
				ConcurrentDocuments: 4,
			},
			expectViolation: true,
			description:     "Should violate when just over limit",
		},
		{
			name: "JustUnderLimit",
			metrics: &ResourceMetrics{
				MemoryUsage:         32*1024*1024 - 1,
				CPUTime:             4999,
				ConcurrentDocuments: 2,
			},
			expectViolation: false,
			description:     "Should not violate when just under limit",
		},
		{
			name: "WellOverLimit",
			metrics: &ResourceMetrics{
				MemoryUsage:         64 * 1024 * 1024,
				CPUTime:             10000,
				ConcurrentDocuments: 10,
			},
			expectViolation: true,
			description:     "Should violate when well over limit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report, err := pm.MonitorResourceUsage(ctx, tc.metrics)
			assert.NoError(t, err, "Should monitor resources successfully")
			assert.NotNil(t, report, "Should return monitoring report")

			hasViolations := len(report.Violations) > 0
			assert.Equal(t, tc.expectViolation, hasViolations, tc.description)

			if tc.expectViolation {
				assert.Equal(t, "violations_detected", report.OverallStatus, "Should indicate violations")
			} else {
				assert.Equal(t, "healthy", report.OverallStatus, "Should indicate healthy status")
			}
		})
	}
}