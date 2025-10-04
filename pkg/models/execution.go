package models

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/alec/pkg/contracts"
)

// ExecutionSession represents a single script execution instance with runtime state
type ExecutionSession struct {
	SessionID      string                      `json:"session_id"`
	Script         *Script                     `json:"script"`
	Status         contracts.ExecutionStatus  `json:"status"`
	StartTime      time.Time                   `json:"start_time"`
	EndTime        *time.Time                  `json:"end_time,omitempty"`
	Duration       time.Duration               `json:"duration"`
	ExitCode       *int                        `json:"exit_code,omitempty"`
	Output         []string                    `json:"output"`
	MaxOutputLines int                         `json:"max_output_lines"`
	Context        context.Context             `json:"-"`
	CancelFunc     context.CancelFunc          `json:"-"`
	PID            *int                        `json:"pid,omitempty"`
	ErrorMessage   string                      `json:"error_message,omitempty"`
}

// NewExecutionSession creates a new execution session
func NewExecutionSession(sessionID string, script *Script, maxOutput int) *ExecutionSession {
	ctx, cancel := context.WithCancel(context.Background())

	return &ExecutionSession{
		SessionID:      sessionID,
		Script:         script,
		Status:         contracts.StatusPending,
		StartTime:      time.Now(),
		Output:         make([]string, 0),
		MaxOutputLines: maxOutput,
		Context:        ctx,
		CancelFunc:     cancel,
	}
}

// Start marks the session as running
func (s *ExecutionSession) Start(pid int) {
	s.Status = contracts.StatusRunning
	s.PID = &pid
}

// Complete marks the session as completed
func (s *ExecutionSession) Complete(exitCode int) {
	now := time.Now()
	s.EndTime = &now
	s.Duration = now.Sub(s.StartTime)
	s.ExitCode = &exitCode

	if exitCode == 0 {
		s.Status = contracts.StatusCompleted
	} else {
		s.Status = contracts.StatusFailed
		s.ErrorMessage = fmt.Sprintf("Script exited with code %d", exitCode)
	}

	if s.CancelFunc != nil {
		s.CancelFunc()
	}
}

// Fail marks the session as failed
func (s *ExecutionSession) Fail(err error) {
	now := time.Now()
	s.EndTime = &now
	s.Duration = now.Sub(s.StartTime)
	s.Status = contracts.StatusFailed
	s.ErrorMessage = err.Error()

	if s.CancelFunc != nil {
		s.CancelFunc()
	}
}

// Cancel marks the session as cancelled
func (s *ExecutionSession) Cancel() {
	if s.Status == contracts.StatusRunning {
		now := time.Now()
		s.EndTime = &now
		s.Duration = now.Sub(s.StartTime)
		s.Status = contracts.StatusCancelled
		s.ErrorMessage = "Execution cancelled by user"
	}

	if s.CancelFunc != nil {
		s.CancelFunc()
	}
}

// Timeout marks the session as timed out
func (s *ExecutionSession) Timeout() {
	now := time.Now()
	s.EndTime = &now
	s.Duration = now.Sub(s.StartTime)
	s.Status = contracts.StatusTimeout
	s.ErrorMessage = "Script execution timed out"

	if s.CancelFunc != nil {
		s.CancelFunc()
	}
}

// AddOutput adds a line of output, respecting the buffer limit
func (s *ExecutionSession) AddOutput(line string) {
	if len(s.Output) >= s.MaxOutputLines {
		// Remove oldest line to make room
		s.Output = s.Output[1:]
	}
	s.Output = append(s.Output, line)
}

// IsRunning returns true if the session is currently running
func (s *ExecutionSession) IsRunning() bool {
	return s.Status == contracts.StatusRunning || s.Status == contracts.StatusPending
}

// IsComplete returns true if the session has finished (success or failure)
func (s *ExecutionSession) IsComplete() bool {
	return s.Status == contracts.StatusCompleted ||
		s.Status == contracts.StatusFailed ||
		s.Status == contracts.StatusCancelled ||
		s.Status == contracts.StatusTimeout
}

// GetResult returns the execution result
func (s *ExecutionSession) GetResult() *contracts.ExecutionResult {
	return &contracts.ExecutionResult{
		SessionID:    s.SessionID,
		Script:       contracts.ScriptInfo{
			ID:   s.Script.ID,
			Name: s.Script.Name,
			Path: s.Script.Path,
			Type: s.Script.Type,
		},
		Status:       s.Status,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
		Duration:     s.Duration,
		ExitCode:     s.ExitCode,
		Output:       s.Output,
		ErrorMessage: s.ErrorMessage,
		PID:          s.PID,
	}
}

// Cleanup releases resources associated with the session
func (s *ExecutionSession) Cleanup() {
	if s.CancelFunc != nil {
		s.CancelFunc()
		s.CancelFunc = nil
	}
	s.Context = nil
}