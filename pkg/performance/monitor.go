package performance

import (
	"context"
	"runtime"
	"sync"
	"time"
	"fmt"
	"sort"
)

// Monitor provides performance monitoring capabilities
type Monitor struct {
	metrics    map[string]*Metric
	mu         sync.RWMutex
	startTime  time.Time
	enabled    bool
	thresholds map[string]time.Duration
}

// Metric represents a performance metric
type Metric struct {
	Name         string
	Count        int64
	TotalTime    time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	LastTime     time.Duration
	Errors       int64
	mu           sync.RWMutex
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc        uint64
	TotalAlloc   uint64
	Sys          uint64
	NumGC        uint32
	PauseTotalNs uint64
}

// ResourceUsage represents system resource usage
type ResourceUsage struct {
	Memory      MemoryStats
	Goroutines  int
	CPUPercent  float64
	Timestamp   time.Time
}

// PerformanceReport represents a comprehensive performance report
type PerformanceReport struct {
	Duration     time.Duration
	Metrics      []*Metric
	Memory       MemoryStats
	Bottlenecks  []Bottleneck
	Suggestions  []string
	GeneratedAt  time.Time
}

// Bottleneck represents a performance bottleneck
type Bottleneck struct {
	Operation   string
	Severity    string
	Description string
	Impact      string
	Suggestions []string
}

// NewMonitor creates a new performance monitor
func NewMonitor() *Monitor {
	return &Monitor{
		metrics:    make(map[string]*Metric),
		startTime:  time.Now(),
		enabled:    true,
		thresholds: getDefaultThresholds(),
	}
}

// Enable enables performance monitoring
func (m *Monitor) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables performance monitoring
func (m *Monitor) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// IsEnabled returns whether monitoring is enabled
func (m *Monitor) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// SetThreshold sets a performance threshold for an operation
func (m *Monitor) SetThreshold(operation string, threshold time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thresholds[operation] = threshold
}

// StartOperation starts timing an operation
func (m *Monitor) StartOperation(name string) *OperationTimer {
	if !m.IsEnabled() {
		return &OperationTimer{enabled: false}
	}
	
	return &OperationTimer{
		monitor:   m,
		operation: name,
		startTime: time.Now(),
		enabled:   true,
	}
}

// RecordOperation records a completed operation
func (m *Monitor) RecordOperation(name string, duration time.Duration, err error) {
	if !m.IsEnabled() {
		return
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	metric, exists := m.metrics[name]
	if !exists {
		metric = &Metric{
			Name:    name,
			MinTime: duration,
			MaxTime: duration,
		}
		m.metrics[name] = metric
	}
	
	metric.mu.Lock()
	defer metric.mu.Unlock()
	
	metric.Count++
	metric.TotalTime += duration
	metric.LastTime = duration
	
	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}
	
	if err != nil {
		metric.Errors++
	}
}

// GetMetric returns a specific metric
func (m *Monitor) GetMetric(name string) *Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if metric, exists := m.metrics[name]; exists {
		// Return a copy to avoid race conditions
		metric.mu.RLock()
		defer metric.mu.RUnlock()
		
		return &Metric{
			Name:      metric.Name,
			Count:     metric.Count,
			TotalTime: metric.TotalTime,
			MinTime:   metric.MinTime,
			MaxTime:   metric.MaxTime,
			LastTime:  metric.LastTime,
			Errors:    metric.Errors,
		}
	}
	
	return nil
}

// GetAllMetrics returns all metrics
func (m *Monitor) GetAllMetrics() []*Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics := make([]*Metric, 0, len(m.metrics))
	for _, metric := range m.metrics {
		metrics = append(metrics, m.GetMetric(metric.Name))
	}
	
	// Sort by total time descending
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].TotalTime > metrics[j].TotalTime
	})
	
	return metrics
}

// GetMemoryStats returns current memory statistics
func (m *Monitor) GetMemoryStats() MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return MemoryStats{
		Alloc:        memStats.Alloc,
		TotalAlloc:   memStats.TotalAlloc,
		Sys:          memStats.Sys,
		NumGC:        memStats.NumGC,
		PauseTotalNs: memStats.PauseTotalNs,
	}
}

// GetResourceUsage returns current resource usage
func (m *Monitor) GetResourceUsage() ResourceUsage {
	return ResourceUsage{
		Memory:     m.GetMemoryStats(),
		Goroutines: runtime.NumGoroutine(),
		Timestamp:  time.Now(),
	}
}

// DetectBottlenecks analyzes metrics to detect performance bottlenecks
func (m *Monitor) DetectBottlenecks() []Bottleneck {
	bottlenecks := []Bottleneck{}
	
	m.mu.RLock()
	thresholds := make(map[string]time.Duration)
	for k, v := range m.thresholds {
		thresholds[k] = v
	}
	m.mu.RUnlock()
	
	metrics := m.GetAllMetrics()
	
	for _, metric := range metrics {
		if metric.Count == 0 {
			continue
		}
		
		avgTime := metric.TotalTime / time.Duration(metric.Count)
		threshold, hasThreshold := thresholds[metric.Name]
		
		// Check if operation exceeds threshold
		if hasThreshold && avgTime > threshold {
			severity := "medium"
			if avgTime > threshold*2 {
				severity = "high"
			}
			
			bottleneck := Bottleneck{
				Operation:   metric.Name,
				Severity:    severity,
				Description: fmt.Sprintf("Average time (%v) exceeds threshold (%v)", avgTime, threshold),
				Impact:      calculateImpact(metric, avgTime, threshold),
				Suggestions: generateSuggestions(metric.Name, avgTime, threshold),
			}
			
			bottlenecks = append(bottlenecks, bottleneck)
		}
		
		// Check for high error rates
		if metric.Errors > 0 {
			errorRate := float64(metric.Errors) / float64(metric.Count)
			if errorRate > 0.1 { // 10% error rate
				bottleneck := Bottleneck{
					Operation:   metric.Name,
					Severity:    "high",
					Description: fmt.Sprintf("High error rate: %.1f%% (%d/%d)", errorRate*100, metric.Errors, metric.Count),
					Impact:      "high",
					Suggestions: []string{
						"Investigate error causes",
						"Add error handling and retry logic",
						"Validate input data",
					},
				}
				
				bottlenecks = append(bottlenecks, bottleneck)
			}
		}
		
		// Check for high variance in execution times
		if metric.MaxTime > metric.MinTime*5 { // 5x variance
			bottleneck := Bottleneck{
				Operation:   metric.Name,
				Severity:    "medium",
				Description: fmt.Sprintf("High variance in execution times (min: %v, max: %v)", metric.MinTime, metric.MaxTime),
				Impact:      "medium",
				Suggestions: []string{
					"Investigate causes of performance variance",
					"Add performance profiling",
					"Optimize worst-case scenarios",
				},
			}
			
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}
	
	// Check memory usage
	memStats := m.GetMemoryStats()
	if memStats.Alloc > 100*1024*1024 { // 100MB
		bottleneck := Bottleneck{
			Operation:   "memory_usage",
			Severity:    "medium",
			Description: fmt.Sprintf("High memory usage: %.2f MB", float64(memStats.Alloc)/(1024*1024)),
			Impact:      "medium",
			Suggestions: []string{
				"Implement memory pooling",
				"Add garbage collection optimization",
				"Review data structure efficiency",
			},
		}
		
		bottlenecks = append(bottlenecks, bottleneck)
	}
	
	return bottlenecks
}

// GenerateReport generates a comprehensive performance report
func (m *Monitor) GenerateReport() PerformanceReport {
	return PerformanceReport{
		Duration:    time.Since(m.startTime),
		Metrics:     m.GetAllMetrics(),
		Memory:      m.GetMemoryStats(),
		Bottlenecks: m.DetectBottlenecks(),
		Suggestions: m.generateOptimizationSuggestions(),
		GeneratedAt: time.Now(),
	}
}

// Reset resets all metrics
func (m *Monitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.metrics = make(map[string]*Metric)
	m.startTime = time.Now()
}

// OperationTimer represents a timer for an operation
type OperationTimer struct {
	monitor   *Monitor
	operation string
	startTime time.Time
	enabled   bool
}

// Stop stops the timer and records the operation
func (ot *OperationTimer) Stop() time.Duration {
	return ot.StopWithError(nil)
}

// StopWithError stops the timer and records the operation with an error
func (ot *OperationTimer) StopWithError(err error) time.Duration {
	if !ot.enabled {
		return 0
	}
	
	duration := time.Since(ot.startTime)
	ot.monitor.RecordOperation(ot.operation, duration, err)
	return duration
}

// MonitorFunction wraps a function with performance monitoring
func (m *Monitor) MonitorFunction(name string, fn func() error) error {
	timer := m.StartOperation(name)
	err := fn()
	timer.StopWithError(err)
	return err
}

// MonitorFunctionWithResult wraps a function with performance monitoring and returns a result
func (m *Monitor) MonitorFunctionWithResult(name string, fn func() (interface{}, error)) (interface{}, error) {
	timer := m.StartOperation(name)
	result, err := fn()
	timer.StopWithError(err)
	return result, err
}

// MonitorContext wraps a context with performance monitoring
func (m *Monitor) MonitorContext(ctx context.Context, name string) (context.Context, func()) {
	timer := m.StartOperation(name)
	
	cancelFunc := func() {
		timer.Stop()
	}
	
	return ctx, cancelFunc
}

// Helper functions

func getDefaultThresholds() map[string]time.Duration {
	return map[string]time.Duration{
		"document_creation":  200 * time.Millisecond,
		"document_loading":   100 * time.Millisecond,
		"document_validation": 50 * time.Millisecond,
		"asset_processing":   150 * time.Millisecond,
		"wasm_execution":     300 * time.Millisecond,
		"container_save":     500 * time.Millisecond,
		"container_load":     300 * time.Millisecond,
		"manifest_parsing":   20 * time.Millisecond,
		"security_validation": 100 * time.Millisecond,
	}
}

func calculateImpact(metric *Metric, avgTime, threshold time.Duration) string {
	ratio := float64(avgTime) / float64(threshold)
	
	if ratio > 3.0 {
		return "high"
	} else if ratio > 2.0 {
		return "medium"
	} else {
		return "low"
	}
}

func generateSuggestions(operation string, avgTime, threshold time.Duration) []string {
	suggestions := []string{}
	
	switch operation {
	case "document_creation":
		suggestions = append(suggestions, 
			"Optimize document structure creation",
			"Implement lazy initialization",
			"Use object pooling for document components")
	case "document_loading":
		suggestions = append(suggestions,
			"Implement streaming document loading",
			"Add compression for document content",
			"Use parallel loading for assets")
	case "asset_processing":
		suggestions = append(suggestions,
			"Implement asset caching",
			"Add asset compression",
			"Use lazy loading for non-critical assets")
	case "wasm_execution":
		suggestions = append(suggestions,
			"Optimize WASM module size",
			"Implement WASM module caching",
			"Review WASM memory usage")
	case "container_save", "container_load":
		suggestions = append(suggestions,
			"Implement incremental saves/loads",
			"Add compression for container data",
			"Optimize ZIP operations")
	default:
		suggestions = append(suggestions,
			"Profile the operation to identify bottlenecks",
			"Consider caching frequently used data",
			"Optimize algorithm complexity")
	}
	
	return suggestions
}

func (m *Monitor) generateOptimizationSuggestions() []string {
	suggestions := []string{}
	
	metrics := m.GetAllMetrics()
	memStats := m.GetMemoryStats()
	
	// Analyze metrics for optimization opportunities
	totalOperations := int64(0)
	totalTime := time.Duration(0)
	
	for _, metric := range metrics {
		totalOperations += metric.Count
		totalTime += metric.TotalTime
	}
	
	if totalOperations > 0 {
		avgTimePerOp := totalTime / time.Duration(totalOperations)
		
		if avgTimePerOp > 100*time.Millisecond {
			suggestions = append(suggestions, "Consider implementing operation batching to reduce per-operation overhead")
		}
		
		if len(metrics) > 10 {
			suggestions = append(suggestions, "High number of different operations detected - consider consolidating similar operations")
		}
	}
	
	// Memory-based suggestions
	if memStats.Alloc > 50*1024*1024 { // 50MB
		suggestions = append(suggestions, "High memory usage detected - consider implementing memory pooling")
	}
	
	if memStats.NumGC > 100 {
		suggestions = append(suggestions, "Frequent garbage collection detected - optimize object allocation patterns")
	}
	
	// Goroutine-based suggestions
	if runtime.NumGoroutine() > 100 {
		suggestions = append(suggestions, "High number of goroutines detected - consider using worker pools")
	}
	
	return suggestions
}

// Global monitor instance
var globalMonitor = NewMonitor()

// Global functions for convenience

// StartOperation starts timing a global operation
func StartOperation(name string) *OperationTimer {
	return globalMonitor.StartOperation(name)
}

// RecordOperation records a global operation
func RecordOperation(name string, duration time.Duration, err error) {
	globalMonitor.RecordOperation(name, duration, err)
}

// GetGlobalMonitor returns the global monitor instance
func GetGlobalMonitor() *Monitor {
	return globalMonitor
}

// EnableGlobalMonitoring enables global performance monitoring
func EnableGlobalMonitoring() {
	globalMonitor.Enable()
}

// DisableGlobalMonitoring disables global performance monitoring
func DisableGlobalMonitoring() {
	globalMonitor.Disable()
}

// GenerateGlobalReport generates a global performance report
func GenerateGlobalReport() PerformanceReport {
	return globalMonitor.GenerateReport()
}

// ResetGlobalMonitor resets the global monitor
func ResetGlobalMonitor() {
	globalMonitor.Reset()
}