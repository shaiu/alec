package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shaiu/alec/pkg/contracts"
	"github.com/shaiu/alec/pkg/models"
)

// ScriptExecutorService implements the ScriptExecutor contract
type ScriptExecutorService struct {
	sessions          map[string]*models.ExecutionSession
	sessionsMutex     sync.RWMutex
	securityValidator *SecurityValidator
	config            *models.ExecutionConfig
}

// NewScriptExecutorService creates a new script executor service
func NewScriptExecutorService(securityValidator *SecurityValidator, config *models.ExecutionConfig) *ScriptExecutorService {
	return &ScriptExecutorService{
		sessions:          make(map[string]*models.ExecutionSession),
		securityValidator: securityValidator,
		config:            config,
	}
}

// ExecuteScript starts execution of a script
func (se *ScriptExecutorService) ExecuteScript(ctx context.Context, script contracts.ScriptInfo) (string, error) {
	// Validate script against security policy
	if err := se.securityValidator.ValidateScriptPath(script.Path); err != nil {
		return "", fmt.Errorf("script validation failed: %w", err)
	}

	// Generate unique session ID
	sessionID := uuid.New().String()

	// Create script model
	scriptModel := &models.Script{
		ID:   script.ID,
		Name: script.Name,
		Path: script.Path,
		Type: script.Type,
	}

	// Create execution session
	session := models.NewExecutionSession(sessionID, scriptModel, se.config.MaxOutputSize)

	// Store session
	se.sessionsMutex.Lock()
	se.sessions[sessionID] = session
	se.sessionsMutex.Unlock()

	// Start execution in background
	go se.executeInBackground(ctx, session)

	return sessionID, nil
}

// executeInBackground runs the script execution in a separate goroutine
func (se *ScriptExecutorService) executeInBackground(ctx context.Context, session *models.ExecutionSession) {
	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, se.config.Timeout)
	// Don't defer cancel here - we'll call it in the completion goroutine

	// Determine shell command
	shell := se.getShell()
	var cmd *exec.Cmd

	switch session.Script.Type {
	case "shell":
		// For shell scripts, try to execute directly if executable, otherwise use shell
		info, err := os.Stat(session.Script.Path)
		if err == nil && info.Mode()&0111 != 0 {
			// Script is executable, run it directly
			cmd = exec.CommandContext(execCtx, session.Script.Path)
		} else {
			// Not executable, use shell
			cmd = exec.CommandContext(execCtx, shell, session.Script.Path)
		}
	case "python":
		cmd = exec.CommandContext(execCtx, "python3", session.Script.Path)
	case "node":
		cmd = exec.CommandContext(execCtx, "node", session.Script.Path)
	default:
		session.Fail(fmt.Errorf("unsupported script type: %s", session.Script.Type))
		return
	}

	// Set up output pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		session.Fail(fmt.Errorf("failed to create stdout pipe: %w", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		session.Fail(fmt.Errorf("failed to create stderr pipe: %w", err))
		return
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		session.Fail(fmt.Errorf("failed to start script '%s': %w", session.Script.Path, err))
		return
	}

	// Mark session as running
	session.Start(cmd.Process.Pid)

	// Stream output in separate goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go se.streamOutput(&wg, session, stdout, "stdout")
	go se.streamOutput(&wg, session, stderr, "stderr")

	// Wait for process completion
	go func() {
		defer cancel() // Cancel the context when process completes

		wg.Wait() // Wait for output streams to finish
		err := cmd.Wait()

		if execCtx.Err() == context.DeadlineExceeded {
			session.Timeout()
		} else if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				session.Complete(exitError.ExitCode())
			} else {
				session.Fail(err)
			}
		} else {
			session.Complete(0)
		}
	}()
}

// streamOutput streams output from a reader to the session
func (se *ScriptExecutorService) streamOutput(wg *sync.WaitGroup, session *models.ExecutionSession, reader io.Reader, stream string) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if stream == "stderr" {
			line = "[stderr] " + line
		}
		session.AddOutput(line)
	}
}

// GetExecutionStatus returns current status of execution session
func (se *ScriptExecutorService) GetExecutionStatus(sessionID string) (*contracts.ExecutionResult, error) {
	se.sessionsMutex.RLock()
	session, exists := se.sessions[sessionID]
	se.sessionsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.GetResult(), nil
}

// StreamOutput returns channel for real-time output streaming
func (se *ScriptExecutorService) StreamOutput(sessionID string) (<-chan contracts.OutputLine, error) {
	se.sessionsMutex.RLock()
	session, exists := se.sessions[sessionID]
	se.sessionsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Create output channel
	outputChan := make(chan contracts.OutputLine, 100)

	// Stream existing output first
	go func() {
		defer close(outputChan)

		// Send existing output
		for _, line := range session.Output {
			outputLine := contracts.OutputLine{
				SessionID: sessionID,
				Line:      line,
				Timestamp: time.Now(),
				Stream:    "stdout", // Simplified for now
			}

			select {
			case outputChan <- outputLine:
			case <-session.Context.Done():
				return
			}
		}

		// Continue streaming if session is still running
		// In a real implementation, this would hook into live output streaming
		for session.IsRunning() {
			time.Sleep(100 * time.Millisecond)
			// Check for new output and send it
		}
	}()

	return outputChan, nil
}

// CancelExecution cancels a running script execution
func (se *ScriptExecutorService) CancelExecution(sessionID string) error {
	se.sessionsMutex.RLock()
	session, exists := se.sessions[sessionID]
	se.sessionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if !session.IsRunning() {
		return fmt.Errorf("session is not running: %s", sessionID)
	}

	// Cancel the session
	session.Cancel()

	return nil
}

// GetExecutionHistory returns recent execution results
func (se *ScriptExecutorService) GetExecutionHistory(limit int) ([]contracts.ExecutionResult, error) {
	se.sessionsMutex.RLock()
	defer se.sessionsMutex.RUnlock()

	var results []contracts.ExecutionResult

	// Collect completed sessions
	for _, session := range se.sessions {
		if session.IsComplete() {
			results = append(results, *session.GetResult())
		}
	}

	// Sort by start time (most recent first)
	// In a real implementation, would use proper sorting

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// CleanupSession removes session data and frees resources
func (se *ScriptExecutorService) CleanupSession(sessionID string) error {
	se.sessionsMutex.Lock()
	defer se.sessionsMutex.Unlock()

	session, exists := se.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Cleanup session resources
	session.Cleanup()

	// Remove from sessions map
	delete(se.sessions, sessionID)

	return nil
}

// getShell determines the appropriate shell for the platform
func (se *ScriptExecutorService) getShell() string {
	if se.config.Shell != "" {
		return se.config.Shell
	}

	switch runtime.GOOS {
	case "windows":
		return "cmd"
	default:
		// Try to find bash, fallback to sh
		if _, err := exec.LookPath("bash"); err == nil {
			return "bash"
		}
		return "sh"
	}
}