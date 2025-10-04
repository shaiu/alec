package integration

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/alec/pkg/services"
)

// TestDirectoryScanAndRefresh tests the complete directory scanning workflow
// This integration test verifies that the ScriptDiscovery service can properly
// scan directories, detect changes, and refresh its state
func TestDirectoryScanAndRefresh(t *testing.T) {
	// Create temporary directory structure for testing
	testDir := t.TempDir()
	scriptsDir := filepath.Join(testDir, "scripts")
	subDir := filepath.Join(scriptsDir, "subdir")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create initial test scripts
	initialScripts := map[string]string{
		filepath.Join(scriptsDir, "script1.sh"):     "#!/bin/bash\necho 'script1'",
		filepath.Join(scriptsDir, "script2.py"):     "#!/usr/bin/env python3\nprint('script2')",
		filepath.Join(subDir, "subscript.js"):       "#!/usr/bin/env node\nconsole.log('subscript');",
	}

	for path, content := range initialScripts {
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			t.Fatalf("Failed to create test script %s: %v", path, err)
		}
	}

	// Initialize service registry for testing
	registry, err := services.NewServiceRegistry()
	if err != nil {
		t.Fatalf("Failed to create service registry: %v", err)
	}

	discovery := registry.GetScriptDiscovery()
	ctx := context.Background()

	t.Run("Initial Directory Scan", func(t *testing.T) {
		// Scan the test directory
		results, err := discovery.ScanDirectories(ctx, []string{scriptsDir})
		if err != nil {
			t.Fatalf("Initial scan failed: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("Expected at least one directory result")
		}

		// Count total scripts found
		totalScripts := 0
		for _, dir := range results {
			totalScripts += len(dir.Scripts)
		}

		if totalScripts != len(initialScripts) {
			t.Errorf("Expected %d scripts, found %d", len(initialScripts), totalScripts)
		}

		// Verify script types are detected correctly
		scriptTypes := make(map[string]bool)
		for _, dir := range results {
			for _, script := range dir.Scripts {
				scriptTypes[script.Type] = true
			}
		}

		expectedTypes := []string{"shell", "python", "node"}
		for _, expectedType := range expectedTypes {
			if !scriptTypes[expectedType] {
				t.Errorf("Expected script type %s not found", expectedType)
			}
		}
	})

	t.Run("Directory Structure Validation", func(t *testing.T) {
		results, err := discovery.ScanDirectories(ctx, []string{scriptsDir})
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// All paths should be absolute
		for _, dir := range results {
			if !filepath.IsAbs(dir.Path) {
				t.Errorf("Directory path should be absolute: %s", dir.Path)
			}

			for _, script := range dir.Scripts {
				if !filepath.IsAbs(script.Path) {
					t.Errorf("Script path should be absolute: %s", script.Path)
				}
			}
		}

		// Verify LastScan timestamp is recent
		for _, dir := range results {
			if time.Since(dir.LastScan) > time.Minute {
				t.Errorf("LastScan timestamp seems too old: %v", dir.LastScan)
			}
		}
	})

	t.Run("Add New Script and Refresh", func(t *testing.T) {
		// Add a new script
		newScriptPath := filepath.Join(scriptsDir, "new_script.rb")
		newScriptContent := "#!/usr/bin/env ruby\nputs 'new script'"

		if err := os.WriteFile(newScriptPath, []byte(newScriptContent), 0755); err != nil {
			t.Fatalf("Failed to create new script: %v", err)
		}

		// Wait a bit to ensure file system changes are registered
		time.Sleep(100 * time.Millisecond)

		// Rescan directory
		results, err := discovery.ScanDirectories(ctx, []string{scriptsDir})
		if err != nil {
			t.Fatalf("Rescan failed: %v", err)
		}

		// Count scripts again
		totalScripts := 0
		foundNewScript := false
		for _, dir := range results {
			for _, script := range dir.Scripts {
				totalScripts++
				if script.Path == newScriptPath {
					foundNewScript = true
					if script.Type != "ruby" {
						t.Errorf("New script type detection failed: got %s, want ruby", script.Type)
					}
				}
			}
		}

		if totalScripts != len(initialScripts)+1 {
			t.Errorf("Expected %d scripts after adding new one, found %d", len(initialScripts)+1, totalScripts)
		}

		if !foundNewScript {
			t.Error("New script was not detected during rescan")
		}
	})

	t.Run("Remove Script and Refresh", func(t *testing.T) {
		// Remove one of the initial scripts
		scriptToRemove := filepath.Join(scriptsDir, "script2.py")
		if err := os.Remove(scriptToRemove); err != nil {
			t.Fatalf("Failed to remove test script: %v", err)
		}

		// Wait a bit for file system changes
		time.Sleep(100 * time.Millisecond)

		// Rescan directory
		results, err := discovery.ScanDirectories(ctx, []string{scriptsDir})
		if err != nil {
			t.Fatalf("Rescan after removal failed: %v", err)
		}

		// Verify the script is no longer found
		for _, dir := range results {
			for _, script := range dir.Scripts {
				if script.Path == scriptToRemove {
					t.Error("Removed script was still found during rescan")
				}
			}
		}
	})

	t.Run("Individual Script Refresh", func(t *testing.T) {
		scriptPath := filepath.Join(scriptsDir, "script1.sh")

		// Test refreshing an existing script
		scriptInfo, err := discovery.RefreshScript(scriptPath)
		if err != nil {
			t.Fatalf("RefreshScript failed for existing script: %v", err)
		}

		if scriptInfo == nil {
			t.Fatal("RefreshScript returned nil for existing script")
		}

		if scriptInfo.Path != scriptPath {
			t.Errorf("RefreshScript returned wrong path: got %s, want %s", scriptInfo.Path, scriptPath)
		}

		// Test refreshing a non-existent script
		nonExistentPath := filepath.Join(scriptsDir, "nonexistent.sh")
		_, err = discovery.RefreshScript(nonExistentPath)
		if err == nil {
			t.Error("RefreshScript should fail for non-existent script")
		}
	})

	t.Run("Permission Handling", func(t *testing.T) {
		// Create a script without execute permissions
		nonExecScript := filepath.Join(scriptsDir, "non_exec.sh")
		if err := os.WriteFile(nonExecScript, []byte("#!/bin/bash\necho 'non-executable'"), 0644); err != nil {
			t.Fatalf("Failed to create non-executable script: %v", err)
		}

		scriptInfo, err := discovery.RefreshScript(nonExecScript)
		if err != nil {
			t.Fatalf("RefreshScript failed for non-executable script: %v", err)
		}

		if scriptInfo.IsExecutable {
			t.Error("Non-executable script was marked as executable")
		}

		// Make it executable and refresh
		if err := os.Chmod(nonExecScript, 0755); err != nil {
			t.Fatalf("Failed to make script executable: %v", err)
		}

		scriptInfo, err = discovery.RefreshScript(nonExecScript)
		if err != nil {
			t.Fatalf("RefreshScript failed after making script executable: %v", err)
		}

		if !scriptInfo.IsExecutable {
			t.Error("Executable script was not marked as executable")
		}
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		// Create a context with very short timeout
		shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// This should fail due to context cancellation
		_, err := discovery.ScanDirectories(shortCtx, []string{scriptsDir})
		if err == nil {
			t.Error("ScanDirectories should fail with cancelled context")
		}

		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Errorf("Expected context cancellation error, got: %v", err)
		}
	})

	t.Run("Large Directory Handling", func(t *testing.T) {
		// Create a directory with many scripts to test performance
		largeDir := filepath.Join(testDir, "large")
		if err := os.MkdirAll(largeDir, 0755); err != nil {
			t.Fatalf("Failed to create large test directory: %v", err)
		}

		// Create 50 test scripts
		for i := 0; i < 50; i++ {
			scriptPath := filepath.Join(largeDir, filepath.Join("script", string(rune('0'+i%10)), ".sh"))
			scriptDir := filepath.Dir(scriptPath)
			if err := os.MkdirAll(scriptDir, 0755); err != nil {
				t.Fatalf("Failed to create script directory: %v", err)
			}

			content := "#!/bin/bash\necho 'script " + string(rune('0'+i%10)) + "'"
			if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
				t.Fatalf("Failed to create test script %d: %v", i, err)
			}
		}

		// Measure scan time
		start := time.Now()
		results, err := discovery.ScanDirectories(ctx, []string{largeDir})
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Large directory scan failed: %v", err)
		}

		// Should complete within reasonable time (less than 5 seconds for 50 files)
		if duration > 5*time.Second {
			t.Errorf("Large directory scan took too long: %v", duration)
		}

		totalScripts := 0
		for _, dir := range results {
			totalScripts += len(dir.Scripts)
		}

		if totalScripts != 50 {
			t.Errorf("Expected 50 scripts in large directory, found %d", totalScripts)
		}
	})
}

// TestDirectoryScanSecurity tests security aspects of directory scanning
func TestDirectoryScanSecurity(t *testing.T) {
	registry, err := services.NewServiceRegistry()
	if err != nil {
		t.Fatalf("Failed to create service registry: %v", err)
	}

	discovery := registry.GetScriptDiscovery()
	ctx := context.Background()

	t.Run("Path Traversal Prevention", func(t *testing.T) {
		maliciousPaths := []string{
			"../../../etc",
			"..\\..\\..\\Windows\\System32",
			"/etc",
			"C:\\Windows\\System32",
		}

		for _, path := range maliciousPaths {
			t.Run("malicious_path_"+path, func(t *testing.T) {
				// Directory scanning should either reject malicious paths or
				// safely handle them without escaping the allowed directories
				results, err := discovery.ScanDirectories(ctx, []string{path})

				// If no error, verify results don't contain sensitive paths
				if err == nil {
					for _, dir := range results {
						if containsSensitivePath(dir.Path) {
							t.Errorf("Scan result contains sensitive path: %s", dir.Path)
						}
						for _, script := range dir.Scripts {
							if containsSensitivePath(script.Path) {
								t.Errorf("Script result contains sensitive path: %s", script.Path)
							}
						}
					}
				}
				// If error, that's also acceptable as the path should be rejected
			})
		}
	})

	t.Run("Symlink Handling", func(t *testing.T) {
		testDir := t.TempDir()

		// Create a script and a symlink to it
		originalScript := filepath.Join(testDir, "original.sh")
		if err := os.WriteFile(originalScript, []byte("#!/bin/bash\necho 'original'"), 0755); err != nil {
			t.Fatalf("Failed to create original script: %v", err)
		}

		symlinkScript := filepath.Join(testDir, "symlink.sh")
		if err := os.Symlink(originalScript, symlinkScript); err != nil {
			// Skip test if symlinks are not supported (e.g., Windows without admin rights)
			t.Skipf("Symlinks not supported: %v", err)
		}

		results, err := discovery.ScanDirectories(ctx, []string{testDir})
		if err != nil {
			t.Fatalf("Scan with symlinks failed: %v", err)
		}

		// Should handle symlinks gracefully (either include them or safely ignore them)
		scriptCount := 0
		for _, dir := range results {
			scriptCount += len(dir.Scripts)
		}

		// Should find at least the original script
		if scriptCount == 0 {
			t.Error("Should find at least the original script")
		}
	})
}

// Helper function to check if a path contains sensitive directories
func containsSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/etc/passwd",
		"/etc/shadow",
		"C:\\Windows\\System32",
		"/root",
		"/var/log",
	}

	for _, sensitive := range sensitivePaths {
		if filepath.Clean(path) == filepath.Clean(sensitive) {
			return true
		}
	}
	return false
}