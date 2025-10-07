package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shaiu/alec/pkg/parser"
)

// TestParseRealScripts tests parsing of actual script files
func TestParseRealScripts(t *testing.T) {
	// Get the directory of this test file
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// If we're in the tests directory, use relative path
	// If we're in the repo root, use tests/fixtures/scripts
	fixturesDir := "./fixtures/scripts"
	if !strings.HasSuffix(cwd, "/tests") {
		fixturesDir = "./tests/fixtures/scripts"
	}

	tests := []struct {
		name           string
		filename       string
		scriptType     string
		wantDescPrefix string
		wantInterp     string
		minLines       int
	}{
		{
			name:           "backup shell script",
			filename:       "backup.sh",
			scriptType:     "shell",
			wantDescPrefix: "Performs automated backups",
			wantInterp:     "/bin/bash",
			minLines:       20,
		},
		{
			name:           "deploy python script",
			filename:       "deploy.py",
			scriptType:     "python",
			wantDescPrefix: "Automated deployment script",
			wantInterp:     "/usr/bin/env python3",
			minLines:       30,
		},
		{
			name:           "simple shell script",
			filename:       "simple.sh",
			scriptType:     "shell",
			wantDescPrefix: "Simple hello world",
			wantInterp:     "/bin/bash",
			minLines:       3,
		},
		{
			name:           "quick test python script",
			filename:       "quick_test.py",
			scriptType:     "python",
			wantDescPrefix: "Quick test runner",
			wantInterp:     "/usr/bin/env python3",
			minLines:       5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(fixturesDir, tt.filename)
			config := parser.DefaultParseConfig()

			metadata, err := parser.ParseScript(scriptPath, tt.scriptType, config)
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}

			// Check description starts with expected prefix
			if !strings.HasPrefix(metadata.Description, tt.wantDescPrefix) {
				t.Errorf("Description = %q, want prefix %q", metadata.Description, tt.wantDescPrefix)
			}

			// Check interpreter
			if metadata.Interpreter != tt.wantInterp {
				t.Errorf("Interpreter = %q, want %q", metadata.Interpreter, tt.wantInterp)
			}

			// Check minimum line count
			if metadata.LineCount < tt.minLines {
				t.Errorf("LineCount = %d, want at least %d", metadata.LineCount, tt.minLines)
			}

			// Check that preview content is not empty
			if metadata.FullContent == "" {
				t.Error("FullContent should not be empty")
			}

			// Check that preview lines is set correctly
			if metadata.PreviewLines == 0 {
				t.Error("PreviewLines should be > 0")
			}

			descPreview := metadata.Description
			if len(descPreview) > 50 {
				descPreview = descPreview[:50]
			}
			t.Logf("✓ Parsed %s: %d lines, description: %q", tt.filename, metadata.LineCount, descPreview)
		})
	}
}
