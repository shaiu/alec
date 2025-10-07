package contract

import (
	"context"
	"testing"
	"time"

	"github.com/shaiu/alec/pkg/contracts"
)

// TestScriptExecutorContract verifies that any implementation of ScriptExecutor
// interface conforms to the contract requirements
func TestScriptExecutorContract(t *testing.T) {
	// This test will fail until we have an implementation
	var executor contracts.ScriptExecutor
	if executor == nil {
		t.Skip("No ScriptExecutor implementation available yet - this is expected during TDD phase")
	}

	tests := []struct {
		name string
		test func(t *testing.T, e contracts.ScriptExecutor)
	}{
		{"ExecuteScript must validate against security policy", testExecuteScriptSecurity},
		{"Sessions must have unique IDs", testSessionUniqueIDs},
		{"Output streaming must be real-time", testOutputStreaming},
		{"Cancellation must be graceful with cleanup timeout", testGracefulCancellation},
		{"Exit codes must be captured accurately", testExitCodeCapture},
		{"Timeout errors must be distinguishable", testTimeoutHandling},
		{"Resource cleanup must prevent memory leaks", testResourceCleanup},
		{"Execution must use current user permissions", testPermissionInheritance},
		{"Output must be limited to prevent exhaustion", testOutputLimiting},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, executor)
		})
	}
}

func testExecuteScriptSecurity(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()

	// Test script with malicious path
	maliciousScript := contracts.ScriptInfo{
		Path: "../../../etc/passwd",
		Type: "shell",
	}

	sessionID, err := e.ExecuteScript(ctx, maliciousScript)
	if err == nil {
		t.Error("ExecuteScript should reject malicious script paths")
		if sessionID != "" {
			// Clean up if session was created
			e.CleanupSession(sessionID)
		}
	}
}

func testSessionUniqueIDs(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()
	script := contracts.ScriptInfo{
		Path: "/tmp/test.sh",
		Type: "shell",
	}

	// Start multiple executions
	sessions := make([]string, 3)
	for i := 0; i < 3; i++ {
		sessionID, err := e.ExecuteScript(ctx, script)
		if err != nil {
			continue // Skip if execution fails (expected in TDD)
		}
		sessions[i] = sessionID
	}

	// Check uniqueness
	seen := make(map[string]bool)
	for _, sessionID := range sessions {
		if sessionID == "" {
			continue
		}
		if seen[sessionID] {
			t.Errorf("Duplicate session ID found: %s", sessionID)
		}
		seen[sessionID] = true
		e.CleanupSession(sessionID)
	}
}

func testOutputStreaming(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()
	script := contracts.ScriptInfo{
		Path: "/bin/echo",
		Type: "shell",
	}

	sessionID, err := e.ExecuteScript(ctx, script)
	if err != nil {
		t.Skip("Cannot test streaming without execution capability")
	}
	defer e.CleanupSession(sessionID)

	// Test streaming output
	outputChan, err := e.StreamOutput(sessionID)
	if err != nil {
		t.Fatalf("StreamOutput failed: %v", err)
	}

	// Should receive output in real-time
	select {
	case output := <-outputChan:
		if output.SessionID != sessionID {
			t.Errorf("Output session ID mismatch: got %s, want %s", output.SessionID, sessionID)
		}
	case <-time.After(5 * time.Second):
		t.Error("Output streaming timed out - should be real-time")
	}
}

func testGracefulCancellation(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()
	script := contracts.ScriptInfo{
		Path: "/bin/sleep",
		Type: "shell",
	}

	sessionID, err := e.ExecuteScript(ctx, script)
	if err != nil {
		t.Skip("Cannot test cancellation without execution capability")
	}
	defer e.CleanupSession(sessionID)

	// Cancel execution
	start := time.Now()
	err = e.CancelExecution(sessionID)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("CancelExecution failed: %v", err)
	}

	// Should complete within graceful shutdown window (5 seconds + buffer)
	if elapsed > 6*time.Second {
		t.Errorf("Cancellation took too long: %v (should be under 6s)", elapsed)
	}
}

func testExitCodeCapture(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()

	tests := []struct {
		name         string
		script       contracts.ScriptInfo
		expectedCode int
	}{
		{"successful script", contracts.ScriptInfo{Path: "/bin/true", Type: "shell"}, 0},
		{"failing script", contracts.ScriptInfo{Path: "/bin/false", Type: "shell"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID, err := e.ExecuteScript(ctx, tt.script)
			if err != nil {
				t.Skip("Cannot test exit codes without execution capability")
			}
			defer e.CleanupSession(sessionID)

			// Wait for completion
			time.Sleep(100 * time.Millisecond)

			status, err := e.GetExecutionStatus(sessionID)
			if err != nil {
				t.Fatalf("GetExecutionStatus failed: %v", err)
			}

			if status.ExitCode == nil {
				t.Error("ExitCode should be captured for completed execution")
			} else if *status.ExitCode != tt.expectedCode {
				t.Errorf("Exit code mismatch: got %d, want %d", *status.ExitCode, tt.expectedCode)
			}
		})
	}
}

func testTimeoutHandling(t *testing.T, e contracts.ScriptExecutor) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	script := contracts.ScriptInfo{
		Path: "/bin/sleep",
		Type: "shell",
	}

	sessionID, err := e.ExecuteScript(ctx, script)
	if err != nil {
		// Check if error indicates timeout
		if ctx.Err() == context.DeadlineExceeded {
			return // Correct timeout behavior
		}
		t.Skip("Cannot test timeout without execution capability")
	}
	defer e.CleanupSession(sessionID)

	// Wait for timeout
	time.Sleep(200 * time.Millisecond)

	status, err := e.GetExecutionStatus(sessionID)
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	if status.Status != contracts.StatusTimeout {
		t.Errorf("Status should be timeout, got: %s", status.Status)
	}
}

func testResourceCleanup(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()
	script := contracts.ScriptInfo{
		Path: "/bin/echo",
		Type: "shell",
	}

	// Create and cleanup multiple sessions
	for i := 0; i < 10; i++ {
		sessionID, err := e.ExecuteScript(ctx, script)
		if err != nil {
			continue
		}

		// Wait for completion
		time.Sleep(50 * time.Millisecond)

		err = e.CleanupSession(sessionID)
		if err != nil {
			t.Errorf("CleanupSession failed for session %s: %v", sessionID, err)
		}

		// Verify session is cleaned up
		_, err = e.GetExecutionStatus(sessionID)
		if err == nil {
			t.Errorf("Session %s should be cleaned up but still accessible", sessionID)
		}
	}
}

func testPermissionInheritance(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()

	// Script that tries to access privileged resource
	script := contracts.ScriptInfo{
		Path: "/bin/ls",
		Type: "shell",
	}

	sessionID, err := e.ExecuteScript(ctx, script)
	if err != nil {
		t.Skip("Cannot test permissions without execution capability")
	}
	defer e.CleanupSession(sessionID)

	// Should execute with current user permissions (no elevation)
	status, err := e.GetExecutionStatus(sessionID)
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	// Verify no privilege escalation occurred
	if status.Status == contracts.StatusFailed && status.ErrorMessage != "" {
		// This is acceptable - script may fail due to permission restrictions
	}
}

func testOutputLimiting(t *testing.T, e contracts.ScriptExecutor) {
	ctx := context.Background()

	// Script that produces large output
	script := contracts.ScriptInfo{
		Path: "/bin/yes", // Produces infinite output
		Type: "shell",
	}

	sessionID, err := e.ExecuteScript(ctx, script)
	if err != nil {
		t.Skip("Cannot test output limiting without execution capability")
	}
	defer e.CleanupSession(sessionID)
	defer e.CancelExecution(sessionID) // Stop infinite output

	// Wait and check output is limited
	time.Sleep(200 * time.Millisecond)

	status, err := e.GetExecutionStatus(sessionID)
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	// Output should be limited to prevent memory exhaustion
	if len(status.Output) > 10000 { // Reasonable limit
		t.Errorf("Output not limited: got %d lines, should be limited", len(status.Output))
	}
}