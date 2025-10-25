// Security event logging implementation

package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileSecurityEventLogger implements SecurityEventLogger using file storage
type FileSecurityEventLogger struct {
	logPath string
	mutex   sync.RWMutex
}

// FileAuditLogger implements AuditLogger using file storage
type FileAuditLogger struct {
	logPath string
	mutex   sync.RWMutex
}

// NewFileSecurityEventLogger creates a new file-based security event logger
func NewFileSecurityEventLogger(logPath string) *FileSecurityEventLogger {
	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		fmt.Printf("Warning: Failed to create log directory: %v\n", err)
	}
	
	return &FileSecurityEventLogger{
		logPath: logPath,
	}
}

// NewFileAuditLogger creates a new file-based audit logger
func NewFileAuditLogger(logPath string) *FileAuditLogger {
	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		fmt.Printf("Warning: Failed to create log directory: %v\n", err)
	}
	
	return &FileAuditLogger{
		logPath: logPath,
	}
}

// LogSecurityEvent logs a security event to file
func (fsel *FileSecurityEventLogger) LogSecurityEvent(event *SecurityEvent) error {
	fsel.mutex.Lock()
	defer fsel.mutex.Unlock()
	
	// Open log file in append mode
	file, err := os.OpenFile(fsel.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open security event log file: %w", err)
	}
	defer file.Close()
	
	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal security event: %w", err)
	}
	
	// Write event with timestamp
	logLine := fmt.Sprintf("%s\n", string(eventJSON))
	if _, err := file.WriteString(logLine); err != nil {
		return fmt.Errorf("failed to write security event: %w", err)
	}
	
	return nil
}

// GetSecurityEvents retrieves security events based on filter
func (fsel *FileSecurityEventLogger) GetSecurityEvents(filter *EventFilter) ([]*SecurityEvent, error) {
	fsel.mutex.RLock()
	defer fsel.mutex.RUnlock()
	
	// Read log file
	file, err := os.Open(fsel.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*SecurityEvent{}, nil
		}
		return nil, fmt.Errorf("failed to open security event log file: %w", err)
	}
	defer file.Close()
	
	var events []*SecurityEvent
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var event SecurityEvent
		if err := decoder.Decode(&event); err != nil {
			continue // Skip malformed entries
		}
		
		// Apply filter
		if fsel.matchesFilter(&event, filter) {
			events = append(events, &event)
		}
	}
	
	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(events) {
			events = events[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(events) {
			events = events[:filter.Limit]
		}
	}
	
	return events, nil
}

// GetEventStatistics returns statistics about security events
func (fsel *FileSecurityEventLogger) GetEventStatistics(timeRange *TimeRange) (*EventStatistics, error) {
	events, err := fsel.GetSecurityEvents(&EventFilter{
		StartTime: &timeRange.Start,
		EndTime:   &timeRange.End,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get events for statistics: %w", err)
	}
	
	stats := &EventStatistics{
		TotalEvents:      len(events),
		EventsByType:     make(map[SecurityEventType]int),
		EventsBySeverity: make(map[SecurityEventSeverity]int),
		EventsByHour:     make(map[string]int),
		TopSources:       []EventSourceStat{},
		TopUsers:         []EventUserStat{},
	}
	
	sourceCount := make(map[string]int)
	userCount := make(map[string]int)
	
	for _, event := range events {
		// Count by type
		stats.EventsByType[event.EventType]++
		
		// Count by severity
		stats.EventsBySeverity[event.Severity]++
		
		// Count by hour
		hour := event.Timestamp.Format("2006-01-02 15")
		stats.EventsByHour[hour]++
		
		// Count sources and users
		sourceCount[event.Source]++
		if event.UserID != "" {
			userCount[event.UserID]++
		}
	}
	
	// Convert maps to sorted slices (simplified - would sort by count in real implementation)
	for source, count := range sourceCount {
		stats.TopSources = append(stats.TopSources, EventSourceStat{Source: source, Count: count})
	}
	
	for user, count := range userCount {
		stats.TopUsers = append(stats.TopUsers, EventUserStat{UserID: user, Count: count})
	}
	
	return stats, nil
}

// LogAuditEvent logs an audit event to file
func (fal *FileAuditLogger) LogAuditEvent(event *AuditEvent) error {
	fal.mutex.Lock()
	defer fal.mutex.Unlock()
	
	// Open log file in append mode
	file, err := os.OpenFile(fal.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer file.Close()
	
	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}
	
	// Write event with timestamp
	logLine := fmt.Sprintf("%s\n", string(eventJSON))
	if _, err := file.WriteString(logLine); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}
	
	return nil
}

// GetAuditTrail retrieves audit events based on filter
func (fal *FileAuditLogger) GetAuditTrail(filter *AuditFilter) ([]*AuditEvent, error) {
	fal.mutex.RLock()
	defer fal.mutex.RUnlock()
	
	// Read log file
	file, err := os.Open(fal.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*AuditEvent{}, nil
		}
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer file.Close()
	
	var events []*AuditEvent
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var event AuditEvent
		if err := decoder.Decode(&event); err != nil {
			continue // Skip malformed entries
		}
		
		// Apply filter
		if fal.matchesAuditFilter(&event, filter) {
			events = append(events, &event)
		}
	}
	
	// Apply limit and offset
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(events) {
			events = events[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(events) {
			events = events[:filter.Limit]
		}
	}
	
	return events, nil
}

// ExportAuditLog exports audit log in specified format
func (fal *FileAuditLogger) ExportAuditLog(format string, timeRange *TimeRange) ([]byte, error) {
	events, err := fal.GetAuditTrail(&AuditFilter{
		StartTime: &timeRange.Start,
		EndTime:   &timeRange.End,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get audit events: %w", err)
	}
	
	switch format {
	case "json":
		return json.MarshalIndent(events, "", "  ")
	case "csv":
		return fal.exportToCSV(events)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// Helper functions

func (fsel *FileSecurityEventLogger) matchesFilter(event *SecurityEvent, filter *EventFilter) bool {
	if filter == nil {
		return true
	}
	
	// Check time range
	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}
	
	// Check event types
	if len(filter.EventTypes) > 0 {
		found := false
		for _, eventType := range filter.EventTypes {
			if event.EventType == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check severities
	if len(filter.Severities) > 0 {
		found := false
		for _, severity := range filter.Severities {
			if event.Severity == severity {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check user ID
	if filter.UserID != "" && event.UserID != filter.UserID {
		return false
	}
	
	// Check policy ID
	if filter.PolicyID != "" && event.PolicyID != filter.PolicyID {
		return false
	}
	
	// Check source
	if filter.Source != "" && event.Source != filter.Source {
		return false
	}
	
	return true
}

func (fal *FileAuditLogger) matchesAuditFilter(event *AuditEvent, filter *AuditFilter) bool {
	if filter == nil {
		return true
	}
	
	// Check time range
	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}
	
	// Check actions
	if len(filter.Actions) > 0 {
		found := false
		for _, action := range filter.Actions {
			if event.Action == action {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check user ID
	if filter.UserID != "" && event.UserID != filter.UserID {
		return false
	}
	
	// Check resource
	if filter.Resource != "" && event.Resource != filter.Resource {
		return false
	}
	
	// Check success
	if filter.Success != nil && event.Success != *filter.Success {
		return false
	}
	
	return true
}

func (fal *FileAuditLogger) exportToCSV(events []*AuditEvent) ([]byte, error) {
	if len(events) == 0 {
		return []byte("timestamp,action,resource,user_id,success,details\n"), nil
	}
	
	csv := "timestamp,action,resource,user_id,success,details\n"
	
	for _, event := range events {
		details, _ := json.Marshal(event.Details)
		line := fmt.Sprintf("%s,%s,%s,%s,%t,\"%s\"\n",
			event.Timestamp.Format(time.RFC3339),
			event.Action,
			event.Resource,
			event.UserID,
			event.Success,
			string(details),
		)
		csv += line
	}
	
	return []byte(csv), nil
}