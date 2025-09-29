// Contract: Script Execution Interface
// This file defines the contract for script execution and output management
// Implementation must pass all associated tests

package contracts

import (
	"context"
	"time"
)

// ExecutionStatus represents the current state of script execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusTimeout   ExecutionStatus = "timeout"
	StatusCancelled ExecutionStatus = "cancelled"
)

// ExecutionResult contains the result of script execution
type ExecutionResult struct {
	SessionID    string          `json:"session_id"`
	Script       ScriptInfo      `json:"script"`
	Status       ExecutionStatus `json:"status"`
	StartTime    time.Time       `json:"start_time"`
	EndTime      *time.Time      `json:"end_time,omitempty"`
	Duration     time.Duration   `json:"duration"`
	ExitCode     *int            `json:"exit_code,omitempty"`
	Output       []string        `json:"output"`
	ErrorMessage string          `json:"error_message,omitempty"`
	PID          *int            `json:"pid,omitempty"`
}

// OutputLine represents a single line of script output
type OutputLine struct {
	SessionID string    `json:"session_id"`
	Line      string    `json:"line"`
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"` // "stdout" or "stderr"
}

// ScriptExecutor interface defines the contract for script execution operations
type ScriptExecutor interface {
	// ExecuteScript starts execution of a script
	// Returns session ID for tracking execution
	// Must handle timeout and cancellation via context
	ExecuteScript(ctx context.Context, script ScriptInfo) (string, error)

	// GetExecutionStatus returns current status of execution session
	// Returns error if session does not exist
	GetExecutionStatus(sessionID string) (*ExecutionResult, error)

	// StreamOutput returns channel for real-time output streaming
	// Channel closes when execution completes or fails
	StreamOutput(sessionID string) (<-chan OutputLine, error)

	// CancelExecution cancels a running script execution
	// Must handle graceful shutdown with fallback to force termination
	CancelExecution(sessionID string) error

	// GetExecutionHistory returns recent execution results
	// Limited to last N executions to prevent memory issues
	GetExecutionHistory(limit int) ([]ExecutionResult, error)

	// CleanupSession removes session data and frees resources
	// Must be called after execution completion
	CleanupSession(sessionID string) error
}

// ExecutionConfig contains configuration for script execution
type ExecutionConfig struct {
	Timeout       time.Duration `json:"timeout"`
	MaxOutputSize int           `json:"max_output_size"`
	Shell         string        `json:"shell"`
	WorkingDir    string        `json:"working_dir"`
	Environment   []string      `json:"environment,omitempty"`
}

// SecurityPolicy defines security constraints for script execution
type SecurityPolicy struct {
	AllowedDirectories []string      `json:"allowed_directories"`
	AllowedExtensions  []string      `json:"allowed_extensions"`
	MaxExecutionTime   time.Duration `json:"max_execution_time"`
	MaxOutputSize      int           `json:"max_output_size"`
	RestrictedCommands []string      `json:"restricted_commands,omitempty"`
}

// Contract Requirements:
// 1. ExecuteScript MUST validate script against security policy
// 2. Sessions MUST have unique IDs (UUID recommended)
// 3. Output streaming MUST be real-time (unbuffered)
// 4. Cancellation MUST be graceful with 5-second cleanup timeout
// 5. Exit codes MUST be captured and reported accurately
// 6. Timeout errors MUST be distinguishable from other failures
// 7. Resource cleanup MUST prevent memory leaks
// 8. Execution MUST use user's current permissions (no privilege escalation)
// 9. Working directory MUST be configurable and validated
// 10. Output MUST be limited to prevent memory exhaustion