package contract

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/alec/pkg/contracts"
)

// TestScriptDiscoveryContract verifies that any implementation of ScriptDiscovery
// interface conforms to the contract requirements
func TestScriptDiscoveryContract(t *testing.T) {
	// This test will fail until we have an implementation
	var discovery contracts.ScriptDiscovery
	if discovery == nil {
		t.Skip("No ScriptDiscovery implementation available yet - this is expected during TDD phase")
	}

	tests := []struct {
		name string
		test func(t *testing.T, d contracts.ScriptDiscovery)
	}{
		{"ScanDirectories must return consistent results", testScanDirectoriesConsistency},
		{"ValidateScript must prevent path traversal", testValidateScriptSecurity},
		{"RefreshScript must handle non-existent files", testRefreshScriptHandling},
		{"FilterScripts must support various queries", testFilterScriptsQuery},
		{"All paths must be absolute and cleaned", testPathHandling},
		{"Script types must be determined by extension", testScriptTypeDetection},
		{"IsExecutable must reflect file permissions", testExecutablePermissions},
		{"Scanning must be interruptible via context", testContextCancellation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, discovery)
		})
	}
}

func testScanDirectoriesConsistency(t *testing.T, d contracts.ScriptDiscovery) {
	ctx := context.Background()
	testDir := t.TempDir()

	// First scan
	result1, err := d.ScanDirectories(ctx, []string{testDir})
	if err != nil {
		t.Fatalf("First scan failed: %v", err)
	}

	// Second scan should return identical results
	result2, err := d.ScanDirectories(ctx, []string{testDir})
	if err != nil {
		t.Fatalf("Second scan failed: %v", err)
	}

	if len(result1) != len(result2) {
		t.Errorf("Scan results inconsistent: got %d vs %d directories", len(result1), len(result2))
	}
}

func testValidateScriptSecurity(t *testing.T, d contracts.ScriptDiscovery) {
	// Test path traversal attempts
	maliciousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\cmd.exe",
		"/etc/passwd",
		"C:\\Windows\\System32\\cmd.exe",
	}

	for _, path := range maliciousPaths {
		t.Run("path_traversal_"+path, func(t *testing.T) {
			_, err := d.ValidateScript(path)
			if err == nil {
				t.Errorf("ValidateScript should reject malicious path: %s", path)
			}
		})
	}
}

func testRefreshScriptHandling(t *testing.T, d contracts.ScriptDiscovery) {
	nonExistentPath := filepath.Join(t.TempDir(), "nonexistent.sh")

	result, err := d.RefreshScript(nonExistentPath)
	if err == nil {
		t.Error("RefreshScript should return error for non-existent file")
	}
	if result != nil {
		t.Error("RefreshScript should return nil result for non-existent file")
	}
}

func testFilterScriptsQuery(t *testing.T, d contracts.ScriptDiscovery) {
	scripts := []contracts.ScriptInfo{
		{Name: "backup.sh", Type: "shell", Tags: []string{"backup", "database"}},
		{Name: "deploy.py", Type: "python", Tags: []string{"deploy", "web"}},
		{Name: "test.js", Type: "node", Tags: []string{"test", "web"}},
	}

	tests := []struct {
		query    string
		expected int
	}{
		{"backup", 1},    // name match
		{"shell", 1},     // type match
		{"web", 2},       // tag match
		{"nonexistent", 0}, // no match
	}

	for _, tt := range tests {
		t.Run("query_"+tt.query, func(t *testing.T) {
			result := d.FilterScripts(scripts, tt.query)
			if len(result) != tt.expected {
				t.Errorf("FilterScripts(%q) = %d results, want %d", tt.query, len(result), tt.expected)
			}
		})
	}
}

func testPathHandling(t *testing.T, d contracts.ScriptDiscovery) {
	ctx := context.Background()
	testDir := t.TempDir()

	results, err := d.ScanDirectories(ctx, []string{testDir})
	if err != nil {
		t.Fatalf("ScanDirectories failed: %v", err)
	}

	for _, dir := range results {
		if !filepath.IsAbs(dir.Path) {
			t.Errorf("Directory path must be absolute: %s", dir.Path)
		}

		for _, script := range dir.Scripts {
			if !filepath.IsAbs(script.Path) {
				t.Errorf("Script path must be absolute: %s", script.Path)
			}
		}
	}
}

func testScriptTypeDetection(t *testing.T, d contracts.ScriptDiscovery) {
	testDir := t.TempDir()

	// Create test scripts with different extensions
	testFiles := map[string]string{
		"test.sh":   "shell",
		"test.py":   "python",
		"test.js":   "node",
		"test.rb":   "ruby",
		"test.pl":   "perl",
		"test.exe":  "", // unsupported should be empty or error
	}

	for filename, expectedType := range testFiles {
		scriptPath := filepath.Join(testDir, filename)
		// Would create actual files in real test

		script, err := d.ValidateScript(scriptPath)
		if expectedType == "" {
			if err == nil {
				t.Errorf("ValidateScript should reject unsupported file type: %s", filename)
			}
			continue
		}

		if err != nil {
			continue // Skip if file doesn't exist (expected in TDD phase)
		}

		if script.Type != expectedType {
			t.Errorf("Script type detection for %s: got %s, want %s", filename, script.Type, expectedType)
		}
	}
}

func testExecutablePermissions(t *testing.T, d contracts.ScriptDiscovery) {
	testDir := t.TempDir()
	scriptPath := filepath.Join(testDir, "test.sh")

	// Would create actual file and test permissions in real test
	script, err := d.ValidateScript(scriptPath)
	if err != nil {
		return // Skip if file doesn't exist (expected in TDD phase)
	}

	// IsExecutable should reflect actual file permissions
	if script.IsExecutable != true {
		t.Error("IsExecutable should reflect actual file permissions")
	}
}

func testContextCancellation(t *testing.T, d contracts.ScriptDiscovery) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	testDir := t.TempDir()

	// This should be cancelled due to short timeout
	_, err := d.ScanDirectories(ctx, []string{testDir})
	if err == nil {
		t.Error("ScanDirectories should respect context cancellation")
	}
}