// Supporting types for security policy management

package security

import (
	"time"

	"github.com/liv-format/liv/pkg/core"
)

// UserContext provides user information for security evaluation
type UserContext struct {
	UserID      string            `json:"user_id"`
	SessionID   string            `json:"session_id"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	Attributes  map[string]string `json:"attributes"`
}

// SecurityEvaluation represents the result of security policy evaluation
type SecurityEvaluation struct {
	DocumentID  string              `json:"document_id"`
	PolicyID    string              `json:"policy_id"`
	EvaluatedAt time.Time           `json:"evaluated_at"`
	UserContext *UserContext        `json:"user_context"`
	Violations  []SecurityViolation `json:"violations"`
	Warnings    []SecurityWarning   `json:"warnings"`
	IsCompliant bool                `json:"is_compliant"`
	Score       int                 `json:"score"` // 0-100 security score
}

// SecurityViolation represents a security policy violation
type SecurityViolation struct {
	Type        string                 `json:"type"`
	Severity    SecurityEventSeverity  `json:"severity"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Remediation string                 `json:"remediation"`
}

// SecurityWarning represents a security warning
type SecurityWarning struct {
	Type           string                 `json:"type"`
	Description    string                 `json:"description"`
	Details        map[string]interface{} `json:"details"`
	Recommendation string                 `json:"recommendation"`
}

// QuarantineRecord represents a quarantined document
type QuarantineRecord struct {
	DocumentID    string           `json:"document_id"`
	PolicyID      string           `json:"policy_id"`
	Reason        string           `json:"reason"`
	QuarantinedAt time.Time        `json:"quarantined_at"`
	ExpiresAt     time.Time        `json:"expires_at"`
	Status        QuarantineStatus `json:"status"`
	ReviewedBy    string           `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time       `json:"reviewed_at,omitempty"`
	ReviewNotes   string           `json:"review_notes,omitempty"`
}

// QuarantineStatus defines quarantine status values
type QuarantineStatus string

const (
	QuarantineStatusActive   QuarantineStatus = "active"
	QuarantineStatusReleased QuarantineStatus = "released"
	QuarantineStatusExpired  QuarantineStatus = "expired"
	QuarantineStatusReviewed QuarantineStatus = "reviewed"
)

// EventFilter defines filters for security event queries
type EventFilter struct {
	StartTime  *time.Time              `json:"start_time,omitempty"`
	EndTime    *time.Time              `json:"end_time,omitempty"`
	EventTypes []SecurityEventType     `json:"event_types,omitempty"`
	Severities []SecurityEventSeverity `json:"severities,omitempty"`
	UserID     string                  `json:"user_id,omitempty"`
	PolicyID   string                  `json:"policy_id,omitempty"`
	Source     string                  `json:"source,omitempty"`
	Limit      int                     `json:"limit,omitempty"`
	Offset     int                     `json:"offset,omitempty"`
}

// AuditFilter defines filters for audit log queries
type AuditFilter struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Actions   []string   `json:"actions,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	Resource  string     `json:"resource,omitempty"`
	Success   *bool      `json:"success,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// TimeRange defines a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// EventStatistics provides statistics about security events
type EventStatistics struct {
	TotalEvents      int                           `json:"total_events"`
	EventsByType     map[SecurityEventType]int     `json:"events_by_type"`
	EventsBySeverity map[SecurityEventSeverity]int `json:"events_by_severity"`
	EventsByHour     map[string]int                `json:"events_by_hour"`
	TopSources       []EventSourceStat             `json:"top_sources"`
	TopUsers         []EventUserStat               `json:"top_users"`
}

// EventSourceStat provides statistics for event sources
type EventSourceStat struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// EventUserStat provides statistics for users
type EventUserStat struct {
	UserID string `json:"user_id"`
	Count  int    `json:"count"`
}

// PolicyTemplate defines a template for creating security policies
type PolicyTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Template    *SystemSecurityPolicy  `json:"template"`
	Variables   map[string]interface{} `json:"variables"`
}

// SecurityMetrics provides security-related metrics
type SecurityMetrics struct {
	TotalPolicies         int            `json:"total_policies"`
	ActiveQuarantines     int            `json:"active_quarantines"`
	ViolationsLast24h     int            `json:"violations_last_24h"`
	ComplianceScore       float64        `json:"compliance_score"`
	ThreatLevel           string         `json:"threat_level"`
	PolicyDistribution    map[string]int `json:"policy_distribution"`
	ViolationsByType      map[string]int `json:"violations_by_type"`
	DocumentsProcessed    int64          `json:"documents_processed"`
	AverageProcessingTime float64        `json:"average_processing_time"`
}

// SystemValidationReport provides system-wide security validation results
type SystemValidationReport struct {
	Timestamp       time.Time               `json:"timestamp"`
	TotalPolicies   int                     `json:"total_policies"`
	Issues          []SystemValidationIssue `json:"issues"`
	Recommendations []string                `json:"recommendations"`
	OverallStatus   string                  `json:"overall_status"`
}

// SystemValidationIssue represents a system configuration issue
type SystemValidationIssue struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	PolicyID       string `json:"policy_id,omitempty"`
	Recommendation string `json:"recommendation"`
}

// ResourceMetrics provides current system resource usage
type ResourceMetrics struct {
	MemoryUsage         int64 `json:"memory_usage"`
	CPUTime             int64 `json:"cpu_time"`
	ConcurrentDocuments int64 `json:"concurrent_documents"`
	NetworkBandwidth    int64 `json:"network_bandwidth"`
	StorageUsage        int64 `json:"storage_usage"`
}

// ResourceMonitoringReport provides resource monitoring results
type ResourceMonitoringReport struct {
	Timestamp     time.Time           `json:"timestamp"`
	Violations    []ResourceViolation `json:"violations"`
	Warnings      []ResourceWarning   `json:"warnings"`
	OverallStatus string              `json:"overall_status"`
}

// ResourceViolation represents a resource limit violation
type ResourceViolation struct {
	PolicyID    string `json:"policy_id"`
	Type        string `json:"type"`
	Current     int64  `json:"current"`
	Limit       int64  `json:"limit"`
	Description string `json:"description"`
}

// ResourceWarning represents a resource usage warning
type ResourceWarning struct {
	PolicyID    string `json:"policy_id"`
	Type        string `json:"type"`
	Current     int64  `json:"current"`
	Threshold   int64  `json:"threshold"`
	Description string `json:"description"`
}

// SecurityConfiguration defines the overall security settings
type SecurityConfiguration struct {
	MaxMemoryPerModule       int64
	MaxCPUTimePerModule      time.Duration
	EnableNetworkAccess      bool
	AllowedDomains           []string
	EnableFileSystem         bool
	MaxConcurrentModules     int
	AuditLogEnabled          bool
	MetricsCollectionEnabled bool
	StrictModeEnabled        bool
	EnableSandbox            bool
	EnableIntegrity          bool
	EnablePermissions        bool
	TrustedDomains           []string
}

// PermissionEvaluationResult represents the result of permission evaluation
type PermissionEvaluationResult struct {
	Allowed     bool
	Violations  []string
	Warnings    []string
	Errors      []string
	Constraints map[string]interface{}
	Timestamp   time.Time
}

// ModuleValidationResult represents WASM module validation result
type ModuleValidationResult struct {
	IsValid    bool
	Valid      bool
	Errors     []string
	Warnings   []string
	ModuleInfo map[string]interface{}
	Timestamp  time.Time
}

// ResourceConstraints defines resource limits
type ResourceConstraints struct {
	MaxMemory       int64
	MaxCPUTime      int64
	MaxThreads      int
	MaxFileSize     int64
	MemoryLimit     int64
	CPUTimeLimit    time.Duration
	ThreadLimit     int
	FileSystemLimit int64
	AllowNetworking bool
	AllowFileSystem bool
}

// SecurityContext provides security context for operations
type SecurityContext struct {
	PolicyID       string
	UserID         string
	SessionID      string
	Permissions    []string
	TrustedDomains []string
	Constraints    *ResourceConstraints
	Policy         *core.SecurityPolicy
	CreatedAt      time.Time
}

// PermissionResponse represents a permission request response
type PermissionResponse struct {
	RequestID string
	Allowed   bool
	Granted   bool
	Reason    string
	Expiry    time.Time
}

// Permission type constants
type PermissionType string

const (
	PermissionTypeMemory     PermissionType = "memory"
	PermissionTypeNetwork    PermissionType = "network"
	PermissionTypeFileSystem PermissionType = "filesystem"
	PermissionTypeImport     PermissionType = "import"
)

// RuntimeMetrics provides real-time runtime metrics
type RuntimeMetrics struct {
	SessionID          string
	ModuleName         string
	MemoryUsage        int64
	CPUTime            int64
	NetworkTraffic     int64
	ActiveModules      int
	Memory             *MemoryUsage
	CPU                *CPUUsage
	Network            *NetworkActivity
	FileSystem         *FileSystemActivity
	PermissionRequests []PermissionRequest
	PolicyViolations   []PolicyViolation
	StartTime          time.Time
	EndTime            time.Time
}

// MemoryUsage tracks memory usage
type MemoryUsage struct {
	Current   int64
	Peak      int64
	Limit     int64
	Allocated int64
	Used      int64
	Timestamp time.Time
}

// CPUUsage tracks CPU usage
type CPUUsage struct {
	TotalTime time.Duration
	Limit     time.Duration
	Used      time.Duration
	Timestamp time.Time
}

// NetworkActivity tracks network activity
type NetworkActivity struct {
	BytesSent       int64
	BytesReceived   int64
	Connections     int
	ConnectionCount int
	RequestCount    int
	LastActivity    time.Time
	Timestamp       time.Time
}

// FileSystemActivity tracks file system activity
type FileSystemActivity struct {
	BytesRead       int64
	BytesWritten    int64
	FilesOpened     int
	ReadOperations  int
	WriteOperations int
	Timestamp       time.Time
}

// PermissionRequest tracks permission requests
type PermissionRequest struct {
	Type           string
	Allowed        bool
	Timestamp      time.Time
	DocumentID     string
	ModuleName     string
	RequestedPerms interface{}
	PolicyID       string
	UserContext    *UserContext
	Justification  string
	RequestedAt    time.Time
}

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	Type        string
	Description string
	Severity    SecuritySeverity
	Timestamp   time.Time
	SessionID   string
	ModuleName  string
	Details     map[string]interface{}
}

// SecuritySeverity represents security severity levels
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
	SecuritySeverityWarning  SecuritySeverity = "warning"
	SecuritySeverityError    SecuritySeverity = "error"
)

// IsMemoryLimitExceeded checks if memory limit is exceeded
func (m *MemoryUsage) IsMemoryLimitExceeded() bool {
	if m == nil {
		return false
	}
	return m.Used > m.Limit
}

// IsCPULimitExceeded checks if CPU limit is exceeded
func (c *CPUUsage) IsCPULimitExceeded() bool {
	if c == nil {
		return false
	}
	return c.Used > c.Limit
}

// GetDuration returns the total execution duration
func (rm *RuntimeMetrics) GetDuration() time.Duration {
	if rm == nil {
		return 0
	}
	return rm.EndTime.Sub(rm.StartTime)
}

// AddPolicyViolation adds a policy violation to the runtime metrics
func (rm *RuntimeMetrics) AddPolicyViolation(violation PolicyViolation) {
	if rm == nil {
		return
	}
	rm.PolicyViolations = append(rm.PolicyViolations, violation)
}

// IsExpired checks if the security context has expired
func (ctx *SecurityContext) IsExpired(maxAge time.Duration) bool {
	if ctx == nil {
		return true
	}
	return time.Since(ctx.CreatedAt) > maxAge
}
